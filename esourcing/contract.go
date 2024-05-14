package esourcing

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/EventStore/EventStore-Client-Go/esdb"
)

type Event interface {
	EventID() string
	EventType() string
	Timestamp() time.Time
	Version() string
	AggregateType() AggregateType
	AggregateID() string
}

type EventType interface{}

type EventTypeRegistry map[string]reflect.Type

type EventMarshaller interface {
	ToEventData(event Event) (esdb.EventData, error)
	FromRecordedEvent(recordedEvent *esdb.RecordedEvent) (Event, error)
}

type EventUpcaster struct {
	From   string
	To     string
	Upcast func(data map[string]interface{}) map[string]interface{}
}

type EventStore interface {
	RegisterEventType(eventType EventType)
	ReadStream(context context.Context, streamID string, options esdb.ReadStreamOptions, count uint64) (events []Event, err error)
	ReadLastEventFromStream(context context.Context, streamID string) (Event, error)
	AppendToStream(context context.Context, streamID string, events []Event) (*esdb.WriteResult, error)
	PersistentSubscribeToStream(ctx context.Context, streamName string, groupName string, options esdb.ConnectToPersistentSubscriptionOptions) (*esdb.PersistentSubscription, error)
	CreatePersistentSubscription(ctx context.Context, streamName string, groupName string, options esdb.PersistentStreamSubscriptionOptions) error
	SubscribeToAll(ctx context.Context, options esdb.SubscribeToAllOptions) (*esdb.Subscription, error)
	SubscribeToStream(ctx context.Context, streamID string, options esdb.SubscribeToStreamOptions) (*esdb.Subscription, error)
	DeleteStream(ctx context.Context, streamID string) error
	GetMarshaller() EventMarshaller
}

type SubscriptionManager interface {
	CreateSubscriptionIfNotExists(subscriptionID string) error
	LastCheckpoint(subscriptionID string) (*esdb.Position, error)
	SaveCheckpoint(subscriptionID string, position *esdb.Position, tx *sql.Tx) error
}
