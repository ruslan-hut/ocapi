package product

import "ocapi/entity"

type Core interface {
	FindProduct(uid string) (interface{}, error)
	LoadProducts(products []*entity.ProductData) error
	LoadProductDescriptions(products []*entity.ProductDescription) error
	LoadProductImages(products []*entity.ProductImage) error
	SetProductImages(products []*entity.ProductData) error
	LoadProductAttributes(products []*entity.ProductAttribute) error
	LoadProductSpecial(products []*entity.ProductSpecial) error
}
