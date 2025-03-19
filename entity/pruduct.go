package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"ocapi/internal/lib/validate"
)

type Product struct {
	Id             int64              `json:"product_id" bson:"product_id" validate:"omitempty"`
	Model          string             `json:"model" bson:"model" validate:"omitempty"`
	Sku            string             `json:"sku" bson:"sku" validate:"omitempty"`
	Upc            string             `json:"upc" bson:"upc" validate:"omitempty"`
	Ean            string             `json:"ean" bson:"ean" validate:"omitempty"`
	Jan            string             `json:"jan" bson:"jan" validate:"omitempty"`
	Isbn           string             `json:"isbn" bson:"isbn" validate:"omitempty"`
	Mpn            string             `json:"mpn" bson:"mpn" validate:"omitempty"`
	Location       string             `json:"location" bson:"location" validate:"omitempty"`
	Quantity       int                `json:"quantity" bson:"quantity" validate:"omitempty"`
	StockStatusId  int                `json:"stock_status_id" bson:"stock_status_id" validate:"omitempty"`
	Image          string             `json:"image" bson:"image" validate:"omitempty"`
	ManufacturerId int                `json:"manufacturer_id" bson:"manufacturer_id" validate:"omitempty"`
	Shipping       int                `json:"shipping" bson:"shipping" validate:"omitempty"`
	Price          float32            `json:"price" bson:"price" validate:"omitempty"`
	Points         int                `json:"points" bson:"points" validate:"omitempty"`
	TaxClassId     int                `json:"tax_class_id" bson:"tax_class_id" validate:"omitempty"`
	DateAvailable  primitive.DateTime `json:"date_available" bson:"date_available" validate:"omitempty"`
	Weight         float32            `json:"weight" bson:"weight" validate:"omitempty"`
	WeightClassId  int                `json:"weight_class_id" bson:"weight_class_id" validate:"omitempty"`
	Length         float32            `json:"length" bson:"length" validate:"omitempty"`
	Width          float32            `json:"width" bson:"width" validate:"omitempty"`
	Height         float32            `json:"height" bson:"height" validate:"omitempty"`
	LengthClassId  int                `json:"length_class_id" bson:"length_class_id" validate:"omitempty"`
	Subtract       int                `json:"subtract" bson:"subtract" validate:"omitempty"`
	Minimum        int                `json:"minimum" bson:"minimum" validate:"omitempty"`
	SortOrder      int                `json:"sort_order" bson:"sort_order" validate:"omitempty"`
	Status         int                `json:"status" bson:"status" validate:"omitempty"`
	DateAdded      primitive.DateTime `json:"date_added" bson:"date_added" validate:"omitempty"`
	DateModified   primitive.DateTime `json:"date_modified" bson:"date_modified" validate:"omitempty"`
	Viewed         int                `json:"viewed" bson:"viewed" validate:"omitempty"`

	Description  string `json:"description" bson:"description" validate:"omitempty"`
	Stock        int    `json:"stock" bson:"stock" validate:"omitempty"`
	Active       bool   `json:"active" bson:"active" validate:"omitempty"`
	Manufacturer string `json:"manufacturer" bson:"manufacturer" validate:"omitempty"`
	Code         string `json:"code" bson:"code" validate:"omitempty"`
	CategoryUUID string `json:"category_uuid" bson:"category_uuid" validate:"omitempty"`
}

func (p *Product) Bind(_ *http.Request) error {
	return validate.Struct(p)
}
