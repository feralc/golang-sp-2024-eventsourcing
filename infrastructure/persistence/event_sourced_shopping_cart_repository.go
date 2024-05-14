package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/entity"
	"github.com/feralc/golang-sp-2024-eventsourcing/domain/repository"
	"github.com/feralc/golang-sp-2024-eventsourcing/esourcing"
	"github.com/google/uuid"
)

var ErrShoppingCartNotFound = fmt.Errorf("cart not found")

type eventSourcedShoppingCartRepository struct {
	eventstore esourcing.EventStore
}

func NewEventSourcedShoppingCartRepository(eventstore esourcing.EventStore) repository.ShoppingCartRepository {
	return &eventSourcedShoppingCartRepository{
		eventstore: eventstore,
	}
}

func (r *eventSourcedShoppingCartRepository) FindByID(ctx context.Context, cartID string) (cart *entity.ShoppingCart, err error) {
	streamID := r.streamID(cartID)

	options := esdb.ReadStreamOptions{
		Direction:      esdb.Forwards,
		From:           esdb.Start{},
		ResolveLinkTos: false,
	}

	var events []esourcing.Event

	storedEvents, err := r.eventstore.ReadStream(ctx, streamID, options, 3000)

	if errors.Is(err, esdb.ErrStreamNotFound) {
		return cart, ErrShoppingCartNotFound
	}

	if err != nil {
		return cart, err
	}

	events = append(events, storedEvents...)

	cart = entity.NewShoppingCart(cartID)
	cart.ClearUncommittedEvents()

	esourcing.RebuildFromEvents(cart, events)

	return cart, nil
}

func (r *eventSourcedShoppingCartRepository) Save(ctx context.Context, cart *entity.ShoppingCart) error {
	uncommitedEvents := cart.UncommittedEvents()

	if len(uncommitedEvents) == 0 {
		return nil
	}

	if len(cart.AggregateID()) == 0 {
		return errors.New("aggregate id cannot be empty")
	}

	streamID := r.streamID(cart.AggregateID())

	_, err := r.eventstore.AppendToStream(ctx, streamID, uncommitedEvents)

	if err != nil {
		return err
	}

	cart.ClearUncommittedEvents()

	return nil
}

func (r *eventSourcedShoppingCartRepository) NextIdentity() string {
	return uuid.NewString()
}

func (r *eventSourcedShoppingCartRepository) streamID(aggregateID string) string {
	return fmt.Sprintf("%s#%s", entity.ShoppingCartAggregateType.String(), aggregateID)
}
