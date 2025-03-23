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
	BatchUID     string  `json:"batch_uid,omitempty"`
}

func (p *ProductData) Bind(_ *http.Request) error {
	return validate.Struct(p)
}

func (p *ProductData) Status() int {
	if p.Active {
		return 1
	}
	return 0
}

func (p *ProductData) StockStatusID() int {
	if p.Quantity > 0 {
		return 7
	}
	return 5
}

type ProductDataRequest struct {
	Full  bool           `json:"full_update"`
	Page  int            `json:"page"`
	Total int            `json:"total"`
	Data  []*ProductData `json:"data" validate:"required,dive"`
}

func (p *ProductDataRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
