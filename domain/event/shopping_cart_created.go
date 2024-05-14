package event

import "github.com/feralc/golang-sp-2024-eventsourcing/esourcing"

type ShoppingCartCreated struct {
	*esourcing.EventBase
	CartID string `json:"cart_id"`
}

func (e ShoppingCartCreated) Version() string {
	return "v1"
}
