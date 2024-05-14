package repository

import (
	"context"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
)

type ShoppingCartRepository interface {
	Save(ctx context.Context, cart *entity.ShoppingCart) error
	FindByID(ctx context.Context, cartID string) (*entity.ShoppingCart, error)
	NextIdentity() string
}
