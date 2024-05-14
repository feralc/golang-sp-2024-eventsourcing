package entity

type ShoppingCartItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

func (item *ShoppingCartItem) Total() float64 {
	return item.Price * float64(item.Quantity)
}
