package projection

import (
	"context"
	"database/sql"
	"log"

	"github.com/feralc/golang-sp-2024-eventsourcing/application/service"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/event"
	"github.com/feralc/golang-sp-2024-eventsourcing/esourcing"
)

type ShoppingCartProjection struct {
	svc        *service.Service
	projection *Projection
}

const projectionName = "shopping-cart-projection"

func NewShoppingCartProjection(svc *service.Service, store esourcing.EventStore) *ShoppingCartProjection {
	return &ShoppingCartProjection{
		svc:        svc,
		projection: NewProjection(svc, store, projectionName),
	}
}

func (p *ShoppingCartProjection) Run(ctx context.Context) {
	log.Println("Shopping cart projection started...")

	p.projection.Run(ctx, []string{
		"ShoppingCartCreated",
		"ShoppingCartItemAdded",
		"ShoppingCartItemRemoved",
		"ShoppingCartCheckedOut",
	}, p.handleShoppingCartEvent)
}

func (p *ShoppingCartProjection) handleShoppingCartEvent(ctx context.Context, evt esourcing.Event, tx *sql.Tx) (err error) {
	switch evt := evt.(type) {
	case event.ShoppingCartCreated:
		err = HandleShoppingCartCreated(tx, evt)
	case event.ShoppingCartItemAdded:
		err = HandleShoppingCartItemAdded(tx, evt)
	case event.ShoppingCartItemRemoved:
		err = HandleShoppingCartItemRemoved(tx, evt)
	case event.ShoppingCartCheckedOut:
		err = HandleShoppingCartCheckedOut(tx, evt)
	}

	return err
}

func HandleShoppingCartCreated(tx *sql.Tx, e event.ShoppingCartCreated) error {
	_, err := tx.Exec("INSERT INTO shopping_cart (cart_id, created_at) VALUES (?,?);",
		e.AggregateID(),
		e.Timestamp(),
	)

	if err != nil {
		return err
	}

	return nil
}

func HandleShoppingCartItemAdded(tx *sql.Tx, e event.ShoppingCartItemAdded) error {
	_, err := tx.Exec(`
		INSERT INTO shopping_cart_item (cart_id, product_id, name, quantity, price, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			quantity = quantity + VALUES(quantity),
			price = VALUES(price),
			created_at = VALUES(created_at);
	`,
		e.AggregateID(),
		e.ProductID,
		e.Name,
		e.Quantity,
		e.Price,
		e.Timestamp(),
	)

	if err != nil {
		return err
	}

	return updateTotal(tx, e.AggregateID())
}

func HandleShoppingCartItemRemoved(tx *sql.Tx, e event.ShoppingCartItemRemoved) error {
	_, err := tx.Exec("DELETE FROM shopping_cart_item WHERE cart_id = ? AND product_id = ?;",
		e.AggregateID(),
		e.ProductID,
	)

	if err != nil {
		return err
	}

	return updateTotal(tx, e.AggregateID())
}

func HandleShoppingCartCheckedOut(tx *sql.Tx, e event.ShoppingCartCheckedOut) error {
	_, err := tx.Exec("DELETE FROM shopping_cart_item WHERE cart_id = ?;",
		e.AggregateID(),
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM shopping_cart WHERE cart_id = ?;",
		e.AggregateID(),
	)

	return err
}

func updateTotal(tx *sql.Tx, cartID string) error {
	var total float64

	row := tx.QueryRow("SELECT COALESCE(SUM(quantity * price), 0) FROM shopping_cart_item WHERE cart_id = ?;", cartID)
	err := row.Scan(&total)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE shopping_cart SET total = ? WHERE cart_id = ?;",
		total,
		cartID,
	)

	return err
}
