package persistence

import (
	"context"
	"errors"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
)

type InMemoryProductRepository struct {
	products map[string]*entity.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: map[string]*entity.Product{
			"123": &entity.Product{
				ProductID: "123",
				Price:     50.5,
			},
			"456": &entity.Product{
				ProductID: "456",
				Price:     22,
			},
			"789": &entity.Product{
				ProductID: "789",
				Price:     105,
			},
			"999": &entity.Product{
				ProductID: "999",
				Price:     230,
			},
		},
	}
}

func (repo *InMemoryProductRepository) FindByID(ctx context.Context, productID string) (*entity.Product, error) {
	product, ok := repo.products[productID]
	if !ok {
		return nil, errors.New("product not found")
	}
	return product, nil
}
