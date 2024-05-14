package service

import (
	"context"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/repository"
)

type ShoppingCartService struct {
	cartRepository    repository.ShoppingCartRepository
	productRepository repository.ProductRepository
}

func NewShoppingCartService(cartRepository repository.ShoppingCartRepository, productRepository repository.ProductRepository) *ShoppingCartService {
	return &ShoppingCartService{
		cartRepository:    cartRepository,
		productRepository: productRepository,
	}
}

func (s *ShoppingCartService) CreateShoppingCart(ctx context.Context) (cartID string, err error) {
	cartID = s.cartRepository.NextIdentity()

	cart := entity.NewShoppingCart(cartID)

	err = s.cartRepository.Save(ctx, cart)

	return cart.CartID(), err
}

func (s *ShoppingCartService) AddItem(ctx context.Context, cartID string, productID string, quantity int) error {
	cart, err := s.cartRepository.FindByID(ctx, cartID)
	if err != nil {
		return err
	}

	product, err := s.productRepository.FindByID(ctx, productID)
	if err != nil {
		return err
	}

	if err := cart.AddItem(product.ProductID, product.Price, quantity); err != nil {
		return err
	}

	return s.cartRepository.Save(ctx, cart)
}

func (s *ShoppingCartService) RemoveItem(ctx context.Context, cartID string, productID string) error {
	cart, err := s.cartRepository.FindByID(ctx, cartID)
	if err != nil {
		return err
	}

	if err := cart.RemoveItem(productID); err != nil {
		return err
	}

	return s.cartRepository.Save(ctx, cart)
}

func (s *ShoppingCartService) Checkout(ctx context.Context, cartID string) error {
	cart, err := s.cartRepository.FindByID(ctx, cartID)
	if err != nil {
		return err
	}

	if err := cart.Checkout(); err != nil {
		return err
	}

	return s.cartRepository.Save(ctx, cart)
}
