package entity

import (
	"fmt"
	"net/http"
	"ocapi/internal/lib/validate"
	"strings"
	"time"
)

// SqlDateTimeLayout is the MySQL DATETIME format used to store special price dates.
const SqlDateTimeLayout = "2006-01-02 15:04:05"

// localDateLayouts are accepted date formats; none of them carries a timezone,
// the value is the shop's wall-clock time.
var localDateLayouts = []string{
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// ParseLocalDate converts a wall-clock date from the accounting system into a DATETIME string
// without any timezone conversion. An empty value, a zero date and the 1C empty date
// (0001-01-01T00:00:00) return "" meaning "no bound".
func ParseLocalDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "0000-00-00") {
		return "", nil
	}
	for _, layout := range localDateLayouts {
		t, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		if t.Year() <= 1 {
			return "", nil
		}
		return t.Format(SqlDateTimeLayout), nil
	}
	return "", fmt.Errorf("invalid date %q: expected YYYY-MM-DDTHH:MM:SS without timezone", value)
}

// SpecialSync is a customer group special price row for a full replacement sync.
type SpecialSync struct {
	ProductUid string  `json:"product_uid" validate:"required"`
	GroupId    int64   `json:"group_id" validate:"required,gt=0"`
	Price      float64 `json:"price" validate:"required,gt=0"`
	Priority   int     `json:"priority" validate:"omitempty"`
	DateStart  string  `json:"date_start"`
	DateEnd    string  `json:"date_end"`
}

// ProductSpecialSyncRequest replaces special prices of the listed groups for every product in Data.
type ProductSpecialSyncRequest struct {
	Data   []*SpecialSync `json:"data" validate:"required,dive"`
	Groups []int64        `json:"groups" validate:"required,min=1,dive,gt=0"`
}

func (p *ProductSpecialSyncRequest) Bind(_ *http.Request) error {
	if err := validate.Struct(p); err != nil {
		return err
	}
	groups := make(map[int64]bool, len(p.Groups))
	for _, group := range p.Groups {
		groups[group] = true
	}
	var err error
	for i, special := range p.Data {
		if !groups[special.GroupId] {
			return fmt.Errorf("data[%d]: group_id %d is not listed in groups", i, special.GroupId)
		}
		if special.DateStart, err = ParseLocalDate(special.DateStart); err != nil {
			return fmt.Errorf("data[%d].date_start: %w", i, err)
		}
		if special.DateEnd, err = ParseLocalDate(special.DateEnd); err != nil {
			return fmt.Errorf("data[%d].date_end: %w", i, err)
		}
	}
	return nil
}

// CustomerSpecial is a personal customer price row.
type CustomerSpecial struct {
	CustomerId int64   `json:"customer_id" validate:"required,min=1"`
	ProductUid string  `json:"product_uid" validate:"required"`
	Price      float64 `json:"price" validate:"required,gt=0"`
	DateStart  string  `json:"date_start"`
	DateEnd    string  `json:"date_end"`
}

// CustomerSpecialSyncRequest replaces personal prices of every customer in Data.
type CustomerSpecialSyncRequest struct {
	Data []*CustomerSpecial `json:"data" validate:"required,dive"`
}

func (c *CustomerSpecialSyncRequest) Bind(_ *http.Request) error {
	if err := validate.Struct(c); err != nil {
		return err
	}
	var err error
	for i, special := range c.Data {
		if special.DateStart, err = ParseLocalDate(special.DateStart); err != nil {
			return fmt.Errorf("data[%d].date_start: %w", i, err)
		}
		if special.DateEnd, err = ParseLocalDate(special.DateEnd); err != nil {
			return fmt.Errorf("data[%d].date_end: %w", i, err)
		}
	}
	return nil
}

// CustomerIdsRequest lists customers whose personal prices are removed.
type CustomerIdsRequest struct {
	Data []int64 `json:"data" validate:"required,min=1,dive,min=1"`
}

func (c *CustomerIdsRequest) Bind(_ *http.Request) error {
	return validate.Struct(c)
}

// SyncResult reports the outcome of a special prices sync.
type SyncResult struct {
	Written          int64    `json:"written"`
	Deleted          int64    `json:"deleted"`
	SkippedProducts  []string `json:"skipped_products"`
	SkippedCustomers []int64  `json:"skipped_customers"`
}

func NewSyncResult() *SyncResult {
	return &SyncResult{
		SkippedProducts:  make([]string, 0),
		SkippedCustomers: make([]int64, 0),
	}
}
