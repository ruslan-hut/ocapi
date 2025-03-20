package entity

type CategoryDescription struct {
	CategoryId      int64  `json:"category_id,omitempty"`
	LanguageId      int64  `json:"language_id,omitempty"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	MetaKeyword     string `json:"meta_keyword,omitempty"`
}

func CategoryDescriptionFromCategoryDescriptionData(category *CategoryDescriptionData) *CategoryDescription {

	return &CategoryDescription{
		CategoryId:      category.CategoryId,
		LanguageId:      category.LanguageId,
		Name:            category.Name,
		Description:     category.Description,
		MetaDescription: category.Name,
		MetaKeyword:     "",
	}
}
