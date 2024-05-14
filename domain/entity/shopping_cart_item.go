package entity

type ShoppingCartItem struct {
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

func (item *ShoppingCartItem) Total() float64 {
	return item.Price * float64(item.Quantity)
}
