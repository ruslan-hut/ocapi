package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductDescription struct {
	ProductUid      string `json:"product_uid" validate:"required"`
	LanguageId      int64  `json:"language_id,omitempty" validate:"required"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaKeyword     string `json:"meta_keyword,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	SeoKeyword      string `json:"seo_keyword,omitempty"`
}

func (p *ProductDescription) Bind(_ *http.Request) error {
	return validate.Struct(p)
}

type ProductDescriptionRequest struct {
	Data []*ProductDescription `json:"data" validate:"required,dive"`
}

func (p *ProductDescriptionRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
