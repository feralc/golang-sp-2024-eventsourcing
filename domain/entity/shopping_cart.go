package entity

import (
	"fmt"

	"github.com/feralc/golang-sp-2024-eventsourcing/domain/event"
	"github.com/feralc/golang-sp-2024-eventsourcing/esourcing"
)

var CartMaxCapacityReachedError = fmt.Errorf("shopping cart max capacity reached")
var CartInvalidQuantityError = fmt.Errorf("shopping cart invalid quantity")
var CartItemNotFoundError = fmt.Errorf("shopping cart item not found")
var CartIsEmptyError = fmt.Errorf("shopping cart is empty")

const CART_CAPACITY = 5

type CartID string

const ShoppingCartAggregateType = esourcing.AggregateType("shopping_cart")

type ShoppingCart struct {
	*esourcing.AggregateRoot
	cartID CartID
	items  []ShoppingCartItem
	total  float64
}

func NewShoppingCart(cartID string) *ShoppingCart {
	cart := &ShoppingCart{
		AggregateRoot: esourcing.NewAggregateRoot(ShoppingCartAggregateType, cartID),
	}

	esourcing.AppendEvent(cart, event.ShoppingCartCreated{
		CartID: cartID,
	})

	return cart
}

func (cart *ShoppingCart) AddItem(productID string, name string, price float64, quantity int) error {
	if len(cart.items)+1 > CART_CAPACITY {
		return CartMaxCapacityReachedError
	}

	if quantity <= 0 {
		return CartInvalidQuantityError
	}

	esourcing.AppendEvent(cart, event.ShoppingCartItemAdded{
		ProductID: productID,
		Name:      name,
		Price:     price,
		Quantity:  quantity,
	})

	return nil
}

func (cart *ShoppingCart) RemoveItem(productID string) error {
	if !cart.HasItem(productID) {
		return CartItemNotFoundError
	}

	esourcing.AppendEvent(cart, event.ShoppingCartItemRemoved{
		ProductID: productID,
	})

	return nil
}

func (cart *ShoppingCart) Checkout() error {
	if cart.IsEmpty() {
		return CartIsEmptyError
	}

	esourcing.AppendEvent(cart, event.ShoppingCartCheckedOut{})

	return nil
}

func (cart *ShoppingCart) CartID() string {
	return string(cart.cartID)
}

func (cart *ShoppingCart) Total() float64 {
	return cart.total
}

func (cart *ShoppingCart) HasItem(productID string) bool {
	return cart.FindItem(productID) != nil
}

func (cart *ShoppingCart) FindItem(productID string) *ShoppingCartItem {
	for idx, item := range cart.items {
		if item.ProductID == productID {
			return &cart.items[idx]
		}
	}
	return nil
}

func (cart *ShoppingCart) Items() []ShoppingCartItem {
	return cart.items
}

func (cart *ShoppingCart) IsEmpty() bool {
	return len(cart.items) == 0
}

func (cart *ShoppingCart) ApplyEvent(e esourcing.Event) {
	switch evt := e.(type) {
	case event.ShoppingCartCreated:
		cart.cartID = CartID(evt.CartID)
		cart.items = []ShoppingCartItem{}

	case event.ShoppingCartItemAdded:
		if existingItem := cart.FindItem(evt.ProductID); existingItem != nil {
			existingItem.Quantity += evt.Quantity
			cart.total += evt.Price * float64(evt.Quantity)
		} else {
			item := ShoppingCartItem{
				ProductID: evt.ProductID,
				Name:      evt.Name,
				Price:     evt.Price,
				Quantity:  evt.Quantity,
			}
			cart.items = append(cart.items, item)
			cart.total += item.Total()
		}

	case event.ShoppingCartItemRemoved:
		for idx, item := range cart.items {
			if item.ProductID == evt.ProductID {
				cart.items = append(cart.items[:idx], cart.items[idx+1:]...)
				cart.total -= item.Total()
			}
		}

	case event.ShoppingCartCheckedOut:
		cart.items = []ShoppingCartItem{}
		cart.total = 0
	}
}
