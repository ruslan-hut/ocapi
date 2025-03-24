package fetch

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type Request struct {
	Table  string `json:"table" validate:"required"`
	Filter string `json:"filter,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Plain  bool   `json:"plain,omitempty"`
}

func (r *Request) Bind(_ *http.Request) error {
	return validate.Struct(r)
}
