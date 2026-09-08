package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
)

// UpdateCustomers applies customer group changes coming from the accounting system.
func (c *Core) UpdateCustomers(data []*entity.Customer) error {
	if c.repo == nil {
		return fmt.Errorf("repository not set")
	}
	if len(data) == 0 {
		return fmt.Errorf("customer data is empty")
	}

	for _, customer := range data {
		err := c.repo.UpdateCustomerGroup(customer.CustomerId, customer.CustomerGroupId)
		if err != nil {
			c.log.Error("updating customer group", sl.Err(err), slog.Int64("customer_id", customer.CustomerId))
			return fmt.Errorf("failed to update customer %d: %w", customer.CustomerId, err)
		}
	}
	return nil
}
