package entity

import (
	"encoding/json"
	"time"
)

// Payment field keys, shared by the database layer and the JSON encoding.
const (
	PaymentFieldStatus   = "status"
	PaymentFieldId       = "id"
	PaymentFieldAmount   = "amount"
	PaymentFieldCurrency = "currency"
	PaymentFieldTime     = "time"
)

// OrderPayment is the payment data of an order read from the custom columns named in
// the order_payment config block. Only configured fields are encoded: a configured
// field with an empty column value is encoded as null, an unconfigured one is omitted,
// so a client can tell "not paid yet" from "this shop does not report it".
type OrderPayment struct {
	OrderId  int64      // set in the payments list, omitted inside an order
	Status   *string    // nil when the column is empty
	Id       *string    // nil when the column is empty
	Amount   *float64   // major units; nil when the column is NULL
	Currency *string    // nil when the column is empty
	Time     *time.Time // nil when the column is NULL
	Fields   []string   // configured field keys, see PaymentField* constants
}

// MarshalJSON encodes order_id (when set) and the configured fields only.
func (p *OrderPayment) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(p.Fields)+1)
	if p.OrderId != 0 {
		m["order_id"] = p.OrderId
	}
	for _, f := range p.Fields {
		switch f {
		case PaymentFieldStatus:
			m[f] = p.Status
		case PaymentFieldId:
			m[f] = p.Id
		case PaymentFieldAmount:
			m[f] = p.Amount
		case PaymentFieldCurrency:
			m[f] = p.Currency
		case PaymentFieldTime:
			m[f] = p.Time
		}
	}
	return json.Marshal(m)
}
