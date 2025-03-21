package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductData struct {
	Uid          string  `json:"product_uid" validate:"required"`
	Article      string  `json:"article,omitempty"`
	Price        float32 `json:"price,omitempty"`
	Quantity     int     `json:"quantity,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	Active       bool    `json:"active,omitempty"`
	CategoryUid  string  `json:"category_uid,omitempty"`
}

func (p *ProductData) Bind(_ *http.Request) error {
	return validate.Struct(p)
}

type ProductDataRequest struct {
	Data []*ProductData `json:"data" validate:"required,dive"`
}

func (p *ProductDataRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
