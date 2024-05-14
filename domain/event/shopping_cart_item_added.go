package event

import "github.com/feralc/golang-sp-2024-eventsourcing/esourcing"

type ShoppingCartItemAdded struct {
	*esourcing.EventBase
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

func (e ShoppingCartItemAdded) Version() string {
	return "v1"
}
