package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type Attribute struct {
	Uid          string                  `json:"attribute_uid" validate:"required"`
	GroupId      int64                   `json:"attribute_group_id" validate:"required"`
	SortOrder    int64                   `json:"sort_order"`
	Descriptions []*AttributeDescription `json:"descriptions" validate:"required,dive"`
}

type AttributeDescription struct {
	LanguageId int64  `json:"language_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
}

type AttributeDataRequest struct {
	Full  bool         `json:"full_update"`
	Page  int          `json:"page"`
	Total int          `json:"total"`
	Data  []*Attribute `json:"data" validate:"required,dive"`
}

func (r *AttributeDataRequest) Bind(_ *http.Request) error {
	return validate.Struct(r)
}
