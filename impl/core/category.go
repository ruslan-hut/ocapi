package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) LoadCategories(categories []*entity.CategoryData) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	err := c.repo.SaveCategories(categories)
	if err != nil {
		return fmt.Errorf("save categories: %w", err)
	}
	return nil
}

func (c *Core) LoadCategoryDescriptions(categories []*entity.CategoryDescriptionData) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	err := c.repo.SaveCategoriesDescription(categories)
	if err != nil {
		return fmt.Errorf("save categories descriptions: %w", err)
	}
	return nil
}
