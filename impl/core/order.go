package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) OrderSearch(id int64) (*entity.Order, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.OrderSearchId(id)
}

func (c *Core) OrderSearchStatus(id int64) ([]int64, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.OrderSearchStatus(id)
}
