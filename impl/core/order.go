package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) OrderSearch(id int64) (*entity.Order, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	order, err := c.repo.OrderSearchId(id)
	if err != nil {
		return nil, err
	}
	products, _ := c.repo.OrderProducts(id)
	if products != nil {
		order.Products = products
	}
	totals, _ := c.repo.OrderTotals(id)
	if totals != nil {
		order.Totals = totals
	}
	return order, nil
}

func (c *Core) OrderSearchStatus(id int64) ([]int64, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.OrderSearchStatus(id)
}

func (c *Core) OrderProducts(id int64) ([]*entity.ProductOrder, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.OrderProducts(id)
}

func (c *Core) OrderSetStatus(id int64, status int, comment string) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	return c.repo.UpdateOrderStatus(id, status, comment)
}
