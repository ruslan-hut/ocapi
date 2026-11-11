package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"ocapi/entity"
	"ocapi/internal/config"
)

// identifierRe limits configured column names to plain identifiers: they are written
// into SQL text, so anything else is rejected rather than escaped.
var identifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)

// paymentColumn is one configured payment field: its JSON key and order-table column.
type paymentColumn struct {
	key    string
	column string
}

// orderPayment is the validated order_payment configuration.
type orderPayment struct {
	columns      []paymentColumn
	minorUnits   bool
	loc          *time.Location
	timeColumn   string
	statusColumn string
}

// fields returns the configured field keys in a stable order.
func (p *orderPayment) fields() []string {
	keys := make([]string, len(p.columns))
	for i, c := range p.columns {
		keys[i] = c.key
	}
	return keys
}

// ConfigureOrderPayment validates the order_payment block against the order table and
// enables payment data. On any error payment data stays disabled (the whole block, so a
// typo cannot silently drop one field) and the error is returned for the caller to log;
// the rest of the service works as without the block.
func (s *MySql) ConfigureOrderPayment(conf config.OrderPayment) error {
	s.payment = nil
	if !conf.Enabled() {
		return nil
	}

	loc := time.UTC
	if conf.TimeLocation != "" {
		l, err := time.LoadLocation(conf.TimeLocation)
		if err != nil {
			return fmt.Errorf("order_payment.time_location: %w", err)
		}
		loc = l
	}

	candidates := []paymentColumn{
		{entity.PaymentFieldStatus, conf.StatusField},
		{entity.PaymentFieldId, conf.IdField},
		{entity.PaymentFieldAmount, conf.AmountField},
		{entity.PaymentFieldCurrency, conf.CurrencyField},
		{entity.PaymentFieldTime, conf.TimeField},
	}

	existing, err := s.orderColumns()
	if err != nil {
		return fmt.Errorf("order_payment: read order columns: %w", err)
	}

	p := &orderPayment{minorUnits: conf.AmountMinorUnits, loc: loc}
	for _, c := range candidates {
		if c.column == "" {
			continue
		}
		if !identifierRe.MatchString(c.column) {
			return fmt.Errorf("order_payment: invalid column name for %s: %q", c.key, c.column)
		}
		if !existing[strings.ToLower(c.column)] {
			return fmt.Errorf("order_payment: column %s not found in %sorder", c.column, s.prefix)
		}
		p.columns = append(p.columns, c)
		switch c.key {
		case entity.PaymentFieldTime:
			p.timeColumn = c.column
		case entity.PaymentFieldStatus:
			p.statusColumn = c.column
		}
	}

	s.payment = p
	return nil
}

// PaymentsListEnabled reports whether payments can be listed by time, which needs the
// time column.
func (s *MySql) PaymentsListEnabled() bool {
	return s.payment != nil && s.payment.timeColumn != ""
}

// orderColumns returns the lower-cased column names of the order table in the current
// database. information_schema is filtered by schema because one server can host several
// shops with identical table names.
func (s *MySql) orderColumns() (map[string]bool, error) {
	rows, err := s.db.Query(
		`SELECT COLUMN_NAME FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?`,
		s.prefix+"order")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	cols := make(map[string]bool)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		cols[strings.ToLower(name)] = true
	}
	return cols, rows.Err()
}

// selectList returns the quoted column list of the configured payment fields.
func (p *orderPayment) selectList() string {
	cols := make([]string, len(p.columns))
	for i, c := range p.columns {
		cols[i] = "`" + c.column + "`"
	}
	return strings.Join(cols, ", ")
}

// scanTargets returns fresh scan destinations, one per configured column.
func (p *orderPayment) scanTargets() []any {
	dest := make([]any, len(p.columns))
	for i, c := range p.columns {
		switch c.key {
		case entity.PaymentFieldAmount:
			dest[i] = &sql.NullFloat64{}
		case entity.PaymentFieldTime:
			dest[i] = &sql.NullTime{}
		default:
			dest[i] = &sql.NullString{}
		}
	}
	return dest
}

// build converts scanned values into an OrderPayment. Empty strings and NULLs become
// nil; the amount is converted to major units; the time column is read as wall clock in
// the configured zone (the driver labels it UTC).
func (p *orderPayment) build(dest []any) *entity.OrderPayment {
	out := &entity.OrderPayment{Fields: p.fields()}
	for i, c := range p.columns {
		switch v := dest[i].(type) {
		case *sql.NullString:
			if !v.Valid || v.String == "" {
				continue
			}
			str := v.String
			switch c.key {
			case entity.PaymentFieldStatus:
				out.Status = &str
			case entity.PaymentFieldId:
				out.Id = &str
			case entity.PaymentFieldCurrency:
				out.Currency = &str
			}
		case *sql.NullFloat64:
			if !v.Valid {
				continue
			}
			amount := v.Float64
			if p.minorUnits {
				amount = amount / 100
			}
			out.Amount = &amount
		case *sql.NullTime:
			if !v.Valid || v.Time.IsZero() {
				continue
			}
			t := v.Time
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), p.loc)
			out.Time = &t
		}
	}
	return out
}

// OrderPayment returns the payment data of an order, or nil when payment data is not
// configured or the order does not exist.
func (s *MySql) OrderPayment(orderId int64) (*entity.OrderPayment, error) {
	p := s.payment
	if p == nil {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT %s FROM %sorder WHERE order_id = ?", p.selectList(), s.prefix)
	dest := p.scanTargets()
	err := s.db.QueryRow(query, orderId).Scan(dest...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("order payment: %w", err)
	}
	return p.build(dest), nil
}

// OrderPayments lists orders whose payment time is after the given moment, ordered by
// time and then order id. afterId continues a page: rows with exactly the from time are
// included only when their order id is greater, so equal times on a page boundary are
// neither lost nor repeated. statuses, when given and the status column is configured,
// filters by payment status.
func (s *MySql) OrderPayments(from time.Time, afterId int64, statuses []string, limit int) ([]*entity.OrderPayment, error) {
	p := s.payment
	if p == nil || p.timeColumn == "" {
		return nil, fmt.Errorf("payments list not configured")
	}
	fromWall := from.In(p.loc).Format("2006-01-02 15:04:05")

	where := fmt.Sprintf("(`%[1]s` > ? OR (`%[1]s` = ? AND order_id > ?))", p.timeColumn)
	args := []any{fromWall, fromWall, afterId}
	if len(statuses) > 0 && p.statusColumn != "" {
		where += fmt.Sprintf(" AND `%s` IN (%s)", p.statusColumn, strings.TrimSuffix(strings.Repeat("?,", len(statuses)), ","))
		for _, st := range statuses {
			args = append(args, st)
		}
	}
	args = append(args, limit)

	query := fmt.Sprintf("SELECT order_id, %s FROM %sorder WHERE %s ORDER BY `%s`, order_id LIMIT ?",
		p.selectList(), s.prefix, where, p.timeColumn)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("order payments: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var result []*entity.OrderPayment
	for rows.Next() {
		var orderId int64
		dest := p.scanTargets()
		if err = rows.Scan(append([]any{&orderId}, dest...)...); err != nil {
			return nil, fmt.Errorf("order payments scan: %w", err)
		}
		payment := p.build(dest)
		payment.OrderId = orderId
		result = append(result, payment)
	}
	return result, rows.Err()
}
