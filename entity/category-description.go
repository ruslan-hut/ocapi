package entity

type CategoryDescription struct {
	CategoryId      int64  `json:"category_id" validate:"required"`
	LanguageId      int64  `json:"language_id" validate:"required"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	MetaKeyword     string `json:"meta_keyword,omitempty"`
}

func CategoryDescriptionFromCategoryDescriptionData(category *CategoryDescriptionData) *CategoryDescription {

	return &CategoryDescription{
		LanguageId:      category.LanguageId,
		Name:            category.Name,
		Description:     category.Description,
		MetaDescription: category.Name,
		MetaKeyword:     "",
	}
}
