package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductData struct {
	UID             string  `json:"uid" validate:"required"`
	Article         string  `json:"article,omitempty"`
	Price           float32 `json:"price,omitempty"`
	Quantity        int     `json:"quantity,omitempty"`
	ManufacturerUID int     `json:"manufacturer_uid,omitempty"`
	Active          bool    `json:"active,omitempty"`
	CategoryUID     int64   `json:"category_uid,omitempty"`
}

func (p *ProductData) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
