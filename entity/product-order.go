package entity

type ProductOrder struct {
	DiscountAmount float32 `json:"discount_amount"`
	DiscountType   string  `json:"discount_type"`
	Ean            string  `json:"ean"`
	Isbn           string  `json:"isbn"`
	Jan            string  `json:"jan"`
	Location       string  `json:"location"`
	Model          string  `json:"model"`
	Mpn            string  `json:"mpn"`
	Name           string  `json:"name"`
	OrderId        int64   `json:"order_id"`
	Price          float32 `json:"price"`
	ProductId      int64   `json:"product_id"`
	Quantity       float32 `json:"quantity"`
	Reward         float32 `json:"reward"`
	Sku            string  `json:"sku"`
	Tax            float32 `json:"tax"`
	Total          float32 `json:"total"`
	Upc            string  `json:"upc"`
	Weight         float32 `json:"weight"`
}
