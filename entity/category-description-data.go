package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryDescriptionData struct {
	CategoryId  int64  `json:"category_id,omitempty"`
	LanguageId  int64  `json:"language_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *CategoryDescriptionData) Bind(_ *http.Request) error {
	return validate.Struct(c)
}
