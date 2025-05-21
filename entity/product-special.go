package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
	"time"
)

type ProductSpecial struct {
	ProductUid string    `json:"product_uid" validate:"required"`
	GroupId    int64     `json:"group_id" validate:"required"`
	Price      float32   `json:"price" validate:"required"`
	Priority   int       `json:"priority"`
	DateStart  time.Time `json:"date_start" validate:"date"`
	DateEnd    time.Time `json:"date_end" validate:"date"`
}

type ProductSpecialRequest struct {
	Data []*ProductSpecial `json:"data" validate:"required,dive"`
}

func (p *ProductSpecialRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
