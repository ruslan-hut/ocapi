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

// customerListLimit caps the page size of the customers listing.
const customerListLimit = 500

// CustomerList returns a page of customers, used to match site customers
// with the accounting system by phone or email.
func (c *Core) CustomerList(limit, offset int) ([]*entity.CustomerInfo, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > customerListLimit {
		limit = customerListLimit
	}
	if offset < 0 {
		offset = 0
	}

	return c.repo.CustomersList(limit, offset)
}
