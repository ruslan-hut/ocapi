package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type CategoryDescriptionData struct {
	CategoryUid string `json:"category_uid" validate:"required"`
	LanguageId  int64  `json:"language_id" validate:"required"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *CategoryDescriptionData) Bind(_ *http.Request) error {
	return validate.Struct(c)
}

type CategoryDescriptionRequest struct {
	Data []*CategoryDescriptionData `json:"data" validate:"required,dive"`
}

func (p *CategoryDescriptionRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
