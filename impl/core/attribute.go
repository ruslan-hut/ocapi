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
	err := c.repo.SaveProductAttributes(attributes)
	if err != nil {
		return err
	}
	return nil
}
