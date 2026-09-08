package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type Customer struct {
	CustomerId      int64 `json:"customer_id" validate:"required,min=1"`
	CustomerGroupId int   `json:"customer_group_id" validate:"required,min=1"`
}

type CustomerData struct {
	Data []*Customer `json:"data" validate:"required,dive"`
}

func (cd *CustomerData) Bind(_ *http.Request) error {
	return validate.Struct(cd)
}
