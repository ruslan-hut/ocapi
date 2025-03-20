package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) FindModel(model string) ([]*entity.Product, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.ProductSearch(model)
}

func (c *Core) LoadProducts(products []*entity.ProductData) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	//c.repo.SaveProducts(products)
	return nil
}
