package order

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type Data struct {
	OrderId       int64  `json:"order_id" validate:"required"`
	OrderStatusId int    `json:"order_status_id" validate:"required"`
	Comment       string `json:"comment,omitempty"`
}

type Request struct {
	Data []*Data `json:"data" validate:"required"`
}

func (r *Request) Bind(_ *http.Request) error {
	return validate.Struct(r)
}
