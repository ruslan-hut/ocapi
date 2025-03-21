package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryData struct {
	CategoryUID string `json:"category_uid" validate:"required"`
	ParentUID   string `json:"parent_uid"`
	SortOrder   int    `json:"sort_order"`
	Menu        bool   `json:"menu"`
	Active      bool   `json:"active"`
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
