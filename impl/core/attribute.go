package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) LoadAttributes(attributes []*entity.Attribute) error {
	if c.repo == nil {
		return fmt.Errorf("repository not set")
	}
	err := c.repo.SaveAttributes(attributes)
	if err != nil {
		return err
	}
	return nil
}

func (c *Core) LoadProductAttributes(attributes []*entity.ProductAttribute) error {
	if c.repo == nil {
		return fmt.Errorf("repository not set")
	}
	if len(attributes) == 0 {
		return nil
	}

	//for _, product := range products {
	//	if err := c.repo.SaveProductAttribute(product); err != nil {
	//		return err
	//	}
	//}

	return nil
}
