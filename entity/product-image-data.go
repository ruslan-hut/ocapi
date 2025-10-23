package entity

type ProductImageData struct {
	ProductUid string `json:"product_uid"`
	FileUid    string `json:"file_uid"`
	SortOrder  int    `json:"sort_order"`
	ImageUrl   string `json:"image"`
}
