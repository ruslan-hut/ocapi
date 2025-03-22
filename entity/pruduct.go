package entity

import (
	"time"
)

type Product struct {
	Id             int64     `json:"product_id" bson:"product_id" validate:"omitempty"`
	Model          string    `json:"model" bson:"model" validate:"omitempty"`
	Sku            string    `json:"sku,omitempty" bson:"sku" validate:"omitempty"`
	Location       string    `json:"location,omitempty" bson:"location" validate:"omitempty"`
	Quantity       int       `json:"quantity" bson:"quantity" validate:"omitempty"`
	StockStatusId  int       `json:"stock_status_id" bson:"stock_status_id" validate:"omitempty"`
	Image          string    `json:"image,omitempty" bson:"image" validate:"omitempty"`
	ManufacturerId int64     `json:"manufacturer_id" bson:"manufacturer_id" validate:"omitempty"`
	Shipping       int       `json:"shipping,omitempty" bson:"shipping" validate:"omitempty"`
	Price          float32   `json:"price" bson:"price" validate:"omitempty"`
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
	Status         int       `json:"status" bson:"status" validate:"omitempty"`
	DateAdded      time.Time `json:"date_added,omitempty" bson:"date_added" validate:"omitempty"`
	DateModified   time.Time `json:"date_modified,omitempty" bson:"date_modified" validate:"omitempty"`
	Viewed         int       `json:"viewed,omitempty" bson:"viewed" validate:"omitempty"`
	BatchUID       string    `json:"batch_uid" bson:"batch_uid" validate:"omitempty"`
}

func ProductFromProductData(product *ProductData) *Product {
	var status = 0
	if product.Active {
		status = 1
	}
	// make product not available if price is 0
	if product.Price == 0 {
		product.Quantity = 0
	}

	var stockStatusId = 5
	if product.Quantity > 0 {
		stockStatusId = 7
	}

	return &Product{
		Model:          product.Uid,
		Sku:            product.Article,
		Quantity:       product.Quantity,
		Minimum:        1,
		Subtract:       1,
		StockStatusId:  stockStatusId,
		DateAvailable:  time.Now().AddDate(0, 0, -3),
		ManufacturerId: 0,
		Shipping:       1,
		Price:          product.Price,
		Points:         0,
		Weight:         0,
		WeightClassId:  1,
		Length:         0,
		Width:          0,
		Height:         0,
		LengthClassId:  1,
		Status:         status,
		TaxClassId:     9,
		SortOrder:      0,
		BatchUID:       product.BatchUID,
		//DateAdded:      time.Now(),
		//DateModified:   time.Now(),
	}
}
