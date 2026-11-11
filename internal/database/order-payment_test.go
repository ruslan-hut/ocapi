package database

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"ocapi/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// testMySql connects to the MySQL named by OCAPI_TEST_MYSQL_DSN (a scratch database,
// e.g. "root:pw@tcp(127.0.0.1:3306)/ocapi_test?parseTime=true") and creates a minimal
// order table with payment columns under the "octest_" prefix. Skipped in -short mode
// and when the variable is unset.
func testMySql(t *testing.T) *MySql {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test")
	}
	dsn := os.Getenv("OCAPI_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OCAPI_TEST_MYSQL_DSN not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	s := &MySql{
		db:         db,
		prefix:     "octest_",
		structure:  make(map[string]map[string]Column),
		statements: make(map[string]*sql.Stmt),
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS octest_order")
		s.Close()
	})
	if _, err = db.Exec("DROP TABLE IF EXISTS octest_order"); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE octest_order (
		order_id INT NOT NULL PRIMARY KEY,
		pay_status VARCHAR(32) NOT NULL DEFAULT '',
		pay_id VARCHAR(64) NOT NULL DEFAULT '',
		pay_amount BIGINT NOT NULL DEFAULT 0,
		pay_currency VARCHAR(3) NOT NULL DEFAULT '',
		pay_time DATETIME NULL DEFAULT NULL)`); err != nil {
		t.Fatalf("create: %v", err)
	}
	return s
}

func fullPaymentConf() config.OrderPayment {
	return config.OrderPayment{
		StatusField:      "pay_status",
		IdField:          "pay_id",
		AmountField:      "pay_amount",
		AmountMinorUnits: true,
		CurrencyField:    "pay_currency",
		TimeField:        "pay_time",
	}
}

func insertPayment(t *testing.T, s *MySql, id int64, status, payId string, amount int64, at any) {
	t.Helper()
	if _, err := s.db.Exec(
		"INSERT INTO octest_order (order_id, pay_status, pay_id, pay_amount, pay_currency, pay_time) VALUES (?, ?, ?, ?, 'PLN', ?)",
		id, status, payId, amount, at); err != nil {
		t.Fatalf("insert %d: %v", id, err)
	}
}

// TestConfigureOrderPayment pins the validation: any bad entry disables the whole
// block, an empty block is "off" without error, and only the time column enables the
// payments list.
func TestConfigureOrderPayment(t *testing.T) {
	s := testMySql(t)

	cases := []struct {
		name       string
		conf       config.OrderPayment
		wantErr    string
		wantOn     bool
		wantListOn bool
	}{
		{name: "empty block", conf: config.OrderPayment{}},
		{name: "full block", conf: fullPaymentConf(), wantOn: true, wantListOn: true},
		{name: "without time", conf: config.OrderPayment{StatusField: "pay_status"}, wantOn: true},
		{name: "column case-insensitive", conf: config.OrderPayment{TimeField: "PAY_TIME"}, wantOn: true, wantListOn: true},
		{name: "missing column", conf: config.OrderPayment{StatusField: "pay_status", IdField: "nope"}, wantErr: "not found"},
		{name: "injection attempt", conf: config.OrderPayment{StatusField: "pay_status`; DROP TABLE x; --"}, wantErr: "invalid column name"},
		{name: "bad location", conf: config.OrderPayment{TimeField: "pay_time", TimeLocation: "Mars/Base"}, wantErr: "time_location"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.ConfigureOrderPayment(tc.conf)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got := s.payment != nil; got != tc.wantOn {
				t.Fatalf("enabled = %v, want %v", got, tc.wantOn)
			}
			if got := s.PaymentsListEnabled(); got != tc.wantListOn {
				t.Fatalf("list enabled = %v, want %v", got, tc.wantListOn)
			}
		})
	}
}

// TestOrderPayment checks the single-order read: configured fields only, empty values
// as nil, amount in major units, time in the configured zone.
func TestOrderPayment(t *testing.T) {
	s := testMySql(t)
	insertPayment(t, s, 1, "paid", "pi_1", 12345, "2026-11-10 09:30:00")
	insertPayment(t, s, 2, "", "", 0, nil)
	if _, err := s.db.Exec("UPDATE octest_order SET pay_currency = '' WHERE order_id = 2"); err != nil {
		t.Fatalf("clear currency: %v", err)
	}

	if p, err := s.OrderPayment(1); err != nil || p != nil {
		t.Fatalf("not configured: got %v, %v; want nil, nil", p, err)
	}

	if err := s.ConfigureOrderPayment(fullPaymentConf()); err != nil {
		t.Fatalf("configure: %v", err)
	}
	p, err := s.OrderPayment(1)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if p.Status == nil || *p.Status != "paid" || p.Id == nil || *p.Id != "pi_1" {
		t.Fatalf("status/id = %v/%v", p.Status, p.Id)
	}
	if p.Amount == nil || *p.Amount != 123.45 {
		t.Fatalf("amount = %v, want 123.45", p.Amount)
	}
	if p.Currency == nil || *p.Currency != "PLN" {
		t.Fatalf("currency = %v", p.Currency)
	}
	want := time.Date(2026, 11, 10, 9, 30, 0, 0, time.UTC)
	if p.Time == nil || !p.Time.Equal(want) {
		t.Fatalf("time = %v, want %v", p.Time, want)
	}

	empty, err := s.OrderPayment(2)
	if err != nil {
		t.Fatalf("read empty: %v", err)
	}
	if empty.Status != nil || empty.Id != nil || empty.Currency != nil || empty.Time != nil {
		t.Fatalf("empty columns must be nil: %+v", empty)
	}
	if empty.Amount == nil || *empty.Amount != 0 {
		t.Fatalf("zero amount must stay 0: %v", empty.Amount)
	}

	if missing, err := s.OrderPayment(999); err != nil || missing != nil {
		t.Fatalf("missing order: got %v, %v; want nil, nil", missing, err)
	}

	// a non-UTC time column is read as wall clock of the configured zone
	conf := fullPaymentConf()
	conf.TimeLocation = "Europe/Warsaw"
	if err = s.ConfigureOrderPayment(conf); err != nil {
		t.Fatalf("configure warsaw: %v", err)
	}
	p, err = s.OrderPayment(1)
	if err != nil {
		t.Fatalf("read warsaw: %v", err)
	}
	if want := time.Date(2026, 11, 10, 8, 30, 0, 0, time.UTC); !p.Time.Equal(want) {
		t.Fatalf("warsaw time = %v, want %v UTC", p.Time.UTC(), want)
	}
}

// TestOrderPayments checks the time-ordered list: strict paging on (time, order id),
// status filter, NULL times excluded.
func TestOrderPayments(t *testing.T) {
	s := testMySql(t)
	insertPayment(t, s, 1, "paid", "pi_1", 100, "2026-11-10 10:00:00")
	insertPayment(t, s, 2, "canceled", "pi_2", 200, "2026-11-10 10:00:00") // same time as 1
	insertPayment(t, s, 3, "paid", "pi_3", 300, "2026-11-10 11:00:00")
	insertPayment(t, s, 4, "requires_capture", "pi_4", 400, nil)       // not final, no time
	insertPayment(t, s, 5, "paid", "pi_5", 500, "2026-11-09 23:00:00") // before from

	if _, err := s.OrderPayments(time.Time{}, 0, nil, 10); err == nil {
		t.Fatal("expected error when not configured")
	}
	if err := s.ConfigureOrderPayment(fullPaymentConf()); err != nil {
		t.Fatalf("configure: %v", err)
	}

	from := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name     string
		from     time.Time
		afterId  int64
		statuses []string
		limit    int
		want     []int64
	}{
		{name: "all after from", from: from, limit: 10, want: []int64{1, 2, 3}},
		{name: "first page", from: from, limit: 1, want: []int64{1}},
		{name: "page boundary on equal time", from: time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC), afterId: 1, limit: 10, want: []int64{2, 3}},
		{name: "strictly after last time", from: time.Date(2026, 11, 10, 11, 0, 0, 0, time.UTC), afterId: 3, limit: 10, want: nil},
		{name: "status filter", from: from, statuses: []string{"paid"}, limit: 10, want: []int64{1, 3}},
		{name: "from in other zone", from: time.Date(2026, 11, 10, 12, 30, 0, 0, time.FixedZone("UTC+2", 2*3600)), limit: 10, want: []int64{3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := s.OrderPayments(tc.from, tc.afterId, tc.statuses, tc.limit)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			var got []int64
			for _, p := range list {
				got = append(got, p.OrderId)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("ids = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("ids = %v, want %v", got, tc.want)
				}
			}
		})
	}
}
