package entity

type ProductImageData struct {
	ProductUid string `json:"product_uid"`
	FileUid    string `json:"file_uid"`
	SortOrder  int    `json:"sort_order"`
	ImageUrl   string `json:"image"`
	IsMain     bool   `json:"is_main"`
}

func NewFromProductImage(productImage *ProductImage) *ProductImageData {
	return &ProductImageData{
		ProductUid: productImage.ProductUid,
		FileUid:    productImage.FileUid,
		IsMain:     productImage.IsMain,
		SortOrder:  productImage.SortOrder,
	}
}
