package core

import "fmt"

func (c *Core) ReadDatabase(table, filter string, limit int) (interface{}, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("database service not available")
	}
	return c.repo.ReadTable(table, filter, limit)
}
