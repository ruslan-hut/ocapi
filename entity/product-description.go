package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductDescription struct {
	ProductUid  string `json:"product_uid" validate:"required"`
	LanguageId  int    `json:"language_id,omitempty" validate:"required"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (p *ProductDescription) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
