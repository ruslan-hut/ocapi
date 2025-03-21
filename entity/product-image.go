package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductImage struct {
	ProductUid string `json:"product_uid" validate:"required"`
	FileUid    string `json:"file_uid" validate:"required"`
	IsMain     bool   `json:"is_main"`
	Version    string `json:"version"`
	FileData   string `json:"file_data" validate:"required,base64"`
}

func (p *ProductImage) Bind(_ *http.Request) error {
	return validate.Struct(p)
}

type ProductImageRequest struct {
	Data []*ProductImage `json:"data" validate:"required,dive"`
}

func (p *ProductImageRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
