package entity

type ProductData struct {
	Name   string `json:"name" validate:"required"`
	NamePl string `json:"name_pl"`
	NameRu string `json:"name_ru"`
	NameEn string `json:"name_en"`

	Description   string `json:"description"`
	DescriptionPl string `json:"description_pl"`
	DescriptionRu string `json:"description_ru"`
	DescriptionEn string `json:"description_en"`

	Uuid         string `json:"uuid" validate:"required"`
	Active       bool   `json:"active"`
	Article      string `json:"article"`
	CategoryUuid string `json:"category_uuid" validate:"required"`
	Stock        string `json:"stock"`
	Currency     string `json:"currency"`
	Price        string `json:"price"`
}
