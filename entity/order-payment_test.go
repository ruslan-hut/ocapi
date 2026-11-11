package entity

import (
	"encoding/json"
	"testing"
	"time"
)

// TestOrderPaymentJSON pins the encoding contract: configured fields appear (null when
// empty), unconfigured ones are omitted, order_id only in the list form.
func TestOrderPaymentJSON(t *testing.T) {
	status := "paid"
	amount := 123.45
	at := time.Date(2026, 11, 10, 9, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		p    *OrderPayment
		want string
	}{
		{
			name: "all fields",
			p: &OrderPayment{Status: &status, Amount: &amount, Time: &at,
				Fields: []string{PaymentFieldStatus, PaymentFieldId, PaymentFieldAmount, PaymentFieldTime}},
			want: `{"amount":123.45,"id":null,"status":"paid","time":"2026-11-10T09:30:00Z"}`,
		},
		{
			name: "only status configured",
			p:    &OrderPayment{Status: &status, Amount: &amount, Fields: []string{PaymentFieldStatus}},
			want: `{"status":"paid"}`,
		},
		{
			name: "list row",
			p:    &OrderPayment{OrderId: 7, Fields: []string{PaymentFieldTime}},
			want: `{"order_id":7,"time":null}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.p)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(b) != tc.want {
				t.Fatalf("got %s, want %s", b, tc.want)
			}
		})
	}

	// inside an order the field is omitted when payment data is off
	b, err := json.Marshal(&Order{OrderID: 1})
	if err != nil {
		t.Fatalf("marshal order: %v", err)
	}
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if _, ok := m["payment"]; ok {
		t.Fatalf("payment must be omitted when nil: %s", b)
	}
}
