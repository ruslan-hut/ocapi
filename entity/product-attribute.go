package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductAttribute struct {
	ProductUid   string `json:"product_uid" validate:"required"`
	AttributeUid string `json:"attribute_uid" validate:"required"`
	LanguageId   int    `json:"language_id" validate:"required"`
	Text         string `json:"text" validate:"required"`
}

type ProductAttributeRequest struct {
	Data []*ProductAttribute `json:"data" validate:"required,dive"`
}

func (p *ProductAttributeRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
