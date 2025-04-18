package product

import "ocapi/entity"

type Core interface {
	FindModel(model string) (interface{}, error)
	LoadProducts(products []*entity.ProductData) error
	LoadProductDescriptions(products []*entity.ProductDescription) error
	LoadProductImages(products []*entity.ProductImage) error
	LoadProductAttributes(products []*entity.ProductAttribute) error
}
