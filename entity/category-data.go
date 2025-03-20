package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryData struct {
	CategoryUID string `json:"category_uid,omitempty" validate:"required"`
	ParentUID   string `json:"parent_uid,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	Top         int    `json:"top,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

func (c *CategoryData) Bind(_ *http.Request) error {
	return validate.Struct(c)
}

func DefaultCategoryData(uid string) *CategoryData {
	return &CategoryData{
		CategoryUID: uid,
		ParentUID:   "",
		SortOrder:   0,
		Top:         0,
		Active:      false,
	}
}
