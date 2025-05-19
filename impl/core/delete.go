package core

import "fmt"

func (c *Core) DeleteFromTable(table, filter string) (int64, error) {
	if c.repo == nil {
		return 0, fmt.Errorf("database service not available")
	}
	return 0, nil //c.repo.DeleteRecords(table, filter)
}
