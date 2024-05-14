package persistence

import (
	"context"
	"errors"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
)

type InMemoryProductRepository struct {
	products []*entity.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	products := []*entity.Product{
		{
			Name:      "Amazing product",
			ProductID: "123",
			Price:     50.5,
		},
		{
			Name:      "Another amazing product",
			ProductID: "456",
			Price:     22,
		},
		{
			Name:      "Awesome product",
			ProductID: "789",
			Price:     105,
		},
		{
			Name:      "Just another product",
			ProductID: "999",
			Price:     230,
		},
	}

	return &InMemoryProductRepository{products: products}
}

func (repo *InMemoryProductRepository) FindByID(ctx context.Context, productID string) (*entity.Product, error) {
	for _, product := range repo.products {
		if product.ProductID == productID {
			return product, nil
		}
	}
	return nil, errors.New("product not found")
}

func (repo *InMemoryProductRepository) All(ctx context.Context) ([]entity.Product, error) {
	products := make([]entity.Product, len(repo.products))
	for i, p := range repo.products {
		products[i] = *p
	}
	return products, nil
}
