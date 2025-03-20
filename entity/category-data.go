package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryData struct {
	CategoryUID int64 `json:"category_uid,omitempty"`
	ParentUID   int64 `json:"parent_uid,omitempty"`
	SortOrder   int   `json:"sort_order,omitempty"`
	Top         int   `json:"top,omitempty"`
	Menu        bool  `json:"menu,omitempty"`
	Active      bool  `json:"active,omitempty"`
}

func (c *CategoryData) Bind(_ *http.Request) error {
	return validate.Struct(c)
}

func DefaultCategoryData(uid int64) *CategoryData {
	return &CategoryData{
		CategoryUID: uid,
		ParentUID:   0,
		SortOrder:   0,
		Top:         0,
		Menu:        false,
		Active:      false,
	}
}
