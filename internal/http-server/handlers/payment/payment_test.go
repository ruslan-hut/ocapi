package payment

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"ocapi/entity"
	"ocapi/impl/core"
)

type fakeCore struct {
	err      error
	from     time.Time
	afterId  int64
	statuses []string
	limit    int
	called   bool
}

func (f *fakeCore) PaymentsList(from time.Time, afterId int64, statuses []string, limit int) ([]*entity.OrderPayment, error) {
	f.called = true
	f.from, f.afterId, f.statuses, f.limit = from, afterId, statuses, limit
	if f.err != nil {
		return nil, f.err
	}
	return nil, nil
}

// TestList checks parameter parsing and status codes of GET /payments.
func TestList(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	from := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name         string
		query        string
		coreErr      error
		wantCode     int
		wantCalled   bool
		wantAfterId  int64
		wantStatuses []string
		wantLimit    int
	}{
		{name: "missing from_date", query: "", wantCode: 400},
		{name: "bad from_date", query: "from_date=2026-11-10", wantCode: 400},
		{name: "bad after_id", query: "from_date=2026-11-10T00:00:00Z&after_id=x", wantCode: 400},
		{name: "bad limit", query: "from_date=2026-11-10T00:00:00Z&limit=-1", wantCode: 400},
		{
			name: "all params", query: "from_date=2026-11-10T00:00:00Z&after_id=42&status=paid,%20canceled,&limit=50",
			wantCode: 200, wantCalled: true, wantAfterId: 42, wantStatuses: []string{"paid", "canceled"}, wantLimit: 50,
		},
		{name: "not configured", query: "from_date=2026-11-10T00:00:00Z", coreErr: core.ErrPaymentsNotConfigured, wantCode: 501, wantCalled: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fc := &fakeCore{err: tc.coreErr}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/payments?"+tc.query, nil)
			List(log, fc).ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("code = %d, want %d; body %s", rec.Code, tc.wantCode, rec.Body)
			}
			if fc.called != tc.wantCalled {
				t.Fatalf("core called = %v, want %v", fc.called, tc.wantCalled)
			}
			if !tc.wantCalled || tc.coreErr != nil {
				return
			}
			if !fc.from.Equal(from) || fc.afterId != tc.wantAfterId || fc.limit != tc.wantLimit ||
				!reflect.DeepEqual(fc.statuses, tc.wantStatuses) {
				t.Fatalf("got from=%v after=%d statuses=%v limit=%d", fc.from, fc.afterId, fc.statuses, fc.limit)
			}
			var body struct {
				Success bool              `json:"success"`
				Data    []json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || !body.Success || body.Data == nil {
				t.Fatalf("body = %s (err %v), want success with an empty list", rec.Body, err)
			}
		})
	}
}
