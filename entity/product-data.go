package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
)

type ProductData struct {
	Uid     string `json:"product_uid" validate:"required"`
	Article string `json:"article"` // goes to the 'model' field
	// use CustomFields to update these values
	//Sku           string         `json:"sku,omitempty" validate:"max=64"`
	//Upc           string         `json:"upc,omitempty" validate:"max=12"`
	//Ean           string         `json:"ean,omitempty" validate:"max=14"`
	//Jan           string         `json:"jan,omitempty" validate:"max=13"`
	//Isbn          string         `json:"isbn,omitempty" validate:"max=17"`
	//Mpn           string         `json:"mpn,omitempty" validate:"max=64"`
	//Location      string         `json:"location,omitempty" validate:"max=128"`
	Price         float64        `json:"price"`
	Quantity      int            `json:"quantity"`
	Manufacturer  string         `json:"manufacturer"`
	Active        bool           `json:"active"`
	Weight        float64        `json:"weight"`
	WeightClassId int            `json:"weight_class_id"`
	Categories    []string       `json:"categories"`
	Attributes    []string       `json:"attributes"`
	Images        []string       `json:"images"`
	CustomFields  []*CustomField `json:"custom_fields" validate:"omitempty,dive"`
	BatchUid      string         `json:"batch_uid"`
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
