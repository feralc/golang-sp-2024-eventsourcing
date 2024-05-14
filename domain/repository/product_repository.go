package repository

import (
	"context"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
)

type ProductRepository interface {
	FindByID(ctx context.Context, productID string) (*entity.Product, error)
}
