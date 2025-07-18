package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type Currency struct {
	Code string  `json:"code" validate:"required,max=3,min=3"`
	Rate float64 `json:"rate" validate:"required,min=0.001"`
}

type CurrencyData struct {
	Data []*Currency `json:"data" validate:"required,dive"`
}

func (cd *CurrencyData) Bind(_ *http.Request) error {
	return validate.Struct(cd)
}
