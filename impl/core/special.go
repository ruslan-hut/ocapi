package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
	"os"
	"path/filepath"
)

// productCachePatterns match OpenCart file cache entries holding product prices:
// product lists and upstore theme blocks (new, popular, category tabs).
var productCachePatterns = []string{
	"cache.product.*",
	"cache.catproductany.*",
	"cache.upstore_productcategorytabs*",
	"cache.theme*",
}

// SyncProductSpecials replaces customer group special prices of the products in the batch.
func (c *Core) SyncProductSpecials(groups []int64, specials []*entity.SpecialSync) (*entity.SyncResult, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	result, err := c.repo.SyncProductSpecials(groups, specials)
	if err != nil {
		return nil, err
	}
	if result.Written > 0 || result.Deleted > 0 {
		c.clearProductCache()
	}
	return result, nil
}

// SyncCustomerSpecials replaces personal prices of the customers in the batch.
// The site applies personal prices after reading its cache, so the cache is not cleared.
func (c *Core) SyncCustomerSpecials(specials []*entity.CustomerSpecial) (*entity.SyncResult, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	return c.repo.SyncCustomerSpecials(specials)
}

// DeleteCustomerSpecials removes all personal prices of the given customers.
func (c *Core) DeleteCustomerSpecials(customerIds []int64) (*entity.SyncResult, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	deleted, err := c.repo.DeleteCustomerSpecials(customerIds)
	if err != nil {
		return nil, err
	}
	result := entity.NewSyncResult()
	result.Deleted = deleted
	return result, nil
}

// clearProductCache removes cached product data from the OpenCart file cache,
// so the site blocks show new prices without waiting for the cache to expire.
func (c *Core) clearProductCache() {
	if c.cachePath == "" {
		return
	}
	var count int
	for _, pattern := range productCachePatterns {
		files, err := filepath.Glob(filepath.Join(c.cachePath, pattern))
		if err != nil {
			c.log.Error("cache files search", sl.Err(err), slog.String("pattern", pattern))
			continue
		}
		for _, file := range files {
			if err = os.Remove(file); err != nil && !os.IsNotExist(err) {
				c.log.Error("removing cache file", sl.Err(err), slog.String("file", file))
				continue
			}
			count++
		}
	}
	c.log.Debug("product cache cleared", slog.Int("files", count))
}
