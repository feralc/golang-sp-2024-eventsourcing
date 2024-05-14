package event

import "github.com/feralc/golang-sp-2024-eventsourcing/esourcing"

type ShoppingCartItemRemoved struct {
	*esourcing.EventBase
	ProductID string `json:"product_id"`
}

func (e ShoppingCartItemRemoved) Version() string {
	return "v1"
}
