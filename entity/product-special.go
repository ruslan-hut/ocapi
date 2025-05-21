package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
	"time"
)

type ProductSpecial struct {
	ProductUid string    `json:"product_uid" validate:"required"`
	GroupId    int64     `json:"group_id" validate:"required,number,gt=0"`
	Price      float32   `json:"price" validate:"required,number,gt=0"`
	Priority   int       `json:"priority" validate:"omitempty,number"`
	DateStart  time.Time `json:"date_start" validate:"omitempty,datetime=2006-01-02T15:04:05"`
	DateEnd    time.Time `json:"date_end" validate:"omitempty,datetime=2006-01-02T15:04:05"`
}

type ProductSpecialRequest struct {
	Data []*ProductSpecial `json:"data" validate:"required,dive"`
}

func (p *ProductSpecialRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
