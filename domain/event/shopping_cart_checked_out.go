package event

import "github.com/feralc/golang-sp-2024-eventsourcing/esourcing"

type ShoppingCartCheckedOut struct {
	*esourcing.EventBase
}

func (e ShoppingCartCheckedOut) Version() string {
	return "v1"
}
