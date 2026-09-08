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

// CustomerInfo is a customer record as exposed by the customers listing.
type CustomerInfo struct {
	CustomerId      int64  `json:"customer_id"`
	CustomerGroupId int    `json:"customer_group_id"`
	FirstName       string `json:"firstname"`
	LastName        string `json:"lastname"`
	Email           string `json:"email"`
	Telephone       string `json:"telephone"`
	Status          int    `json:"status"`
}
