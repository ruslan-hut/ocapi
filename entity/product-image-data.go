package entity

type ProductImageData struct {
	ProductUid string `json:"product_uid"`
	FileUid    string `json:"file_uid"`
	SortOrder  int    `json:"sort_order"`
	ImageUrl   string `json:"image"`
	IsMain     bool   `json:"is_main"`
}

func NewFromProductImage(productImage *ProductImage) *ProductImageData {
	so := int(productImage.SortOrder)
	if so < 0 {
		so = 0
	}
	if so > 255 {
		so = 255
	}
	return &ProductImageData{
		ProductUid: productImage.ProductUid,
		FileUid:    productImage.FileUid,
		IsMain:     productImage.IsMain,
		SortOrder:  so,
	}
}
