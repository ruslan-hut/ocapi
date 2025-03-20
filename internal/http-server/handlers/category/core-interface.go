package category

import "ocapi/entity"

type Core interface {
	LoadCategories(categories []*entity.CategoryData) error
	LoadCategoryDescriptions(categories []*entity.CategoryDescriptionData) error
}
