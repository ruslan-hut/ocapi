package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
	"time"
)

func (c *Core) OrderSearch(id int64) (*entity.Order, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	order, err := c.repo.OrderSearchId(id)
	if err != nil {
		return nil, err
	}

	products, err := c.repo.OrderProducts(id)
	if err != nil {
		c.log.Warn("failed to fetch order products", slog.Int64("order_id", id), sl.Err(err))
	} else {
		order.Products = products
	}

	totals, err := c.repo.OrderTotals(id)
	if err != nil {
		c.log.Warn("failed to fetch order totals", slog.Int64("order_id", id), sl.Err(err))
	} else {
		order.Totals = totals
	}

	return order, nil
}

func (c *Core) OrderSearchStatus(id int64, from time.Time) ([]int64, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.OrderSearchStatus(id, from)
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
