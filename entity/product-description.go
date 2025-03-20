package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductDescription struct {
	ProductId   string  `json:"product_id" validate:"required"`
	LanguageId  string  `json:"language_id,omitempty"`
	Name        float32 `json:"name,omitempty"`
	Description int     `json:"description,omitempty"`
}

func (p *ProductDescription) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
