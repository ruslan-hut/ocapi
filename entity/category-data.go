package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryData struct {
	CategoryUID string `json:"category_uid,omitempty" validate:"required"`
	ParentUID   string `json:"parent_uid,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	Top         int    `json:"top,omitempty" validate:"in=0 1"`
	Active      bool   `json:"active,omitempty"`
}

func (c *CategoryData) Bind(_ *http.Request) error {
	return validate.Struct(c)
}

type CategoryDataRequest struct {
	Data []*CategoryData `json:"data" validate:"required,dive"`
}

func (p *CategoryDataRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
