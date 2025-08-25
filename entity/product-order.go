package entity

type ProductOrder struct {
	DiscountAmount float64 `json:"discount_amount"`
	DiscountType   string  `json:"discount_type"`
	Ean            string  `json:"ean"`
	Isbn           string  `json:"isbn"`
	Jan            string  `json:"jan"`
	Location       string  `json:"location"`
	Model          string  `json:"model"`
	Mpn            string  `json:"mpn"`
	Name           string  `json:"name"`
	OrderId        int64   `json:"order_id"`
	Price          float64 `json:"price"`
	ProductId      int64   `json:"product_id"`
	ProductUid     string  `json:"product_uid"`
	Quantity       float64 `json:"quantity"`
	Reward         float64 `json:"reward"`
	Sku            string  `json:"sku"`
	Tax            float64 `json:"tax"`
	Total          float64 `json:"total"`
	Upc            string  `json:"upc"`
	Weight         float64 `json:"weight"`
}
