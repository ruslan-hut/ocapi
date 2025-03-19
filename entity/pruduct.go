package entity

import (
	"net/http"
	"ocapi/internal/lib/validate"
	"time"
)

type Product struct {
	Id             int64     `json:"product_id" bson:"product_id" validate:"omitempty"`
	Model          string    `json:"model,omitempty" bson:"model" validate:"omitempty"`
	Sku            string    `json:"sku,omitempty" bson:"sku" validate:"omitempty"`
	Upc            string    `json:"upc,omitempty" bson:"upc" validate:"omitempty"`
	Ean            string    `json:"ean,omitempty" bson:"ean" validate:"omitempty"`
	Jan            string    `json:"jan,omitempty" bson:"jan" validate:"omitempty"`
	Isbn           string    `json:"isbn,omitempty" bson:"isbn" validate:"omitempty"`
	Mpn            string    `json:"mpn,omitempty" bson:"mpn" validate:"omitempty"`
	Location       string    `json:"location,omitempty" bson:"location" validate:"omitempty"`
	Quantity       int       `json:"quantity,omitempty" bson:"quantity" validate:"omitempty"`
	StockStatusId  int       `json:"stock_status_id,omitempty" bson:"stock_status_id" validate:"omitempty"`
	Image          string    `json:"image,omitempty" bson:"image" validate:"omitempty"`
	ManufacturerId int       `json:"manufacturer_id,omitempty" bson:"manufacturer_id" validate:"omitempty"`
	Shipping       int       `json:"shipping,omitempty" bson:"shipping" validate:"omitempty"`
	Price          float32   `json:"price,omitempty" bson:"price" validate:"omitempty"`
	Points         int       `json:"points,omitempty" bson:"points" validate:"omitempty"`
	TaxClassId     int       `json:"tax_class_id,omitempty" bson:"tax_class_id" validate:"omitempty"`
	DateAvailable  time.Time `json:"date_available,omitempty" bson:"date_available" validate:"omitempty"`
	Weight         float32   `json:"weight,omitempty" bson:"weight" validate:"omitempty"`
	WeightClassId  int       `json:"weight_class_id,omitempty" bson:"weight_class_id" validate:"omitempty"`
	Length         float32   `json:"length,omitempty" bson:"length" validate:"omitempty"`
	Width          float32   `json:"width,omitempty" bson:"width" validate:"omitempty"`
	Height         float32   `json:"height,omitempty" bson:"height" validate:"omitempty"`
	LengthClassId  int       `json:"length_class_id,omitempty" bson:"length_class_id" validate:"omitempty"`
	Subtract       int       `json:"subtract,omitempty" bson:"subtract" validate:"omitempty"`
	Minimum        int       `json:"minimum,omitempty" bson:"minimum" validate:"omitempty"`
	SortOrder      int       `json:"sort_order,omitempty" bson:"sort_order" validate:"omitempty"`
	Status         int       `json:"status,omitempty" bson:"status" validate:"omitempty"`
	DateAdded      time.Time `json:"date_added,omitempty" bson:"date_added" validate:"omitempty"`
	DateModified   time.Time `json:"date_modified,omitempty" bson:"date_modified" validate:"omitempty"`
	Viewed         int       `json:"viewed,omitempty" bson:"viewed" validate:"omitempty"`

	Description  string `json:"description,omitempty" bson:"description" validate:"omitempty"`
	Stock        int    `json:"stock,omitempty" bson:"stock" validate:"omitempty"`
	Active       bool   `json:"active,omitempty" bson:"active" validate:"omitempty"`
	Manufacturer string `json:"manufacturer,omitempty" bson:"manufacturer" validate:"omitempty"`
	Code         string `json:"code,omitempty" bson:"code" validate:"omitempty"`
	CategoryUUID string `json:"category_uuid,omitempty" bson:"category_uuid" validate:"omitempty"`
}

func (p *Product) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
