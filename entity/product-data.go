package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductData struct {
	Uid           string   `json:"product_uid" validate:"required"`
	Article       string   `json:"article"`
	Price         float64  `json:"price"`
	Quantity      int      `json:"quantity"`
	Manufacturer  string   `json:"manufacturer"`
	Active        bool     `json:"active"`
	Weight        float64  `json:"weight"`
	WeightClassId int      `json:"weight_class_id"`
	Categories    []string `json:"categories"`
	Attributes    []string `json:"attributes"`
	Images        []string `json:"images"`
	BatchUid      string   `json:"batch_uid"`
}

func (p *ProductData) Bind(_ *http.Request) error {
	return validate.Struct(p)
}

func (p *ProductData) Status() int {
	if p.Active {
		return 1
	}
	return 0
}

func (p *ProductData) StockStatusID() int {
	if p.Quantity > 0 {
		return 7
	}
	return 5
}

type ProductDataRequest struct {
	Full  bool           `json:"full_update"`
	Page  int            `json:"page"`
	Total int            `json:"total"`
	Data  []*ProductData `json:"data" validate:"required,dive"`
}

func (p *ProductDataRequest) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
