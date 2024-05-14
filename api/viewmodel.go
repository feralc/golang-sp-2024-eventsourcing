package api

import "github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"

type ShoppingCartItemViewModel struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
}

type ShoppingCartViewModel struct {
	CartID string                      `json:"cart_id"`
	Total  float64                     `json:"total"`
	Items  []ShoppingCartItemViewModel `json:"items"`
}

func NewShoppingCartViewModel(cart *entity.ShoppingCart) ShoppingCartViewModel {
	items := make([]ShoppingCartItemViewModel, len(cart.Items()))

	for i, item := range cart.Items() {
		items[i] = ShoppingCartItemViewModel{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Total:     item.Total(),
		}
	}

	return ShoppingCartViewModel{
		CartID: cart.CartID(),
		Total:  cart.Total(),
		Items:  items,
	}
}
