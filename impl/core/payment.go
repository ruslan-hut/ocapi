package core

import (
	"fmt"
	"ocapi/entity"
	"time"
)

// paymentsMaxLimit caps one page of the payments list.
const paymentsMaxLimit = 500

// ErrPaymentsNotConfigured is returned when the payments list is requested but the
// order_payment block has no time column (or is absent or invalid).
var ErrPaymentsNotConfigured = fmt.Errorf("payments list not configured")

// PaymentsList returns a page of order payments with a payment time after from (see
// Repository.OrderPayments for the afterId paging rule). limit is clamped to
// 1..paymentsMaxLimit.
func (c *Core) PaymentsList(from time.Time, afterId int64, statuses []string, limit int) ([]*entity.OrderPayment, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	if !c.repo.PaymentsListEnabled() {
		return nil, ErrPaymentsNotConfigured
	}
	if limit <= 0 || limit > paymentsMaxLimit {
		limit = paymentsMaxLimit
	}
	return c.repo.OrderPayments(from, afterId, statuses, limit)
}
