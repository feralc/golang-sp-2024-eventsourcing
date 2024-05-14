package esourcing

import (
	"context"
	"fmt"
	"reflect"

	"github.com/EventStore/EventStore-Client-Go/esdb"
)

type eventStore struct {
	client            *esdb.Client
	eventTypeRegistry EventTypeRegistry
	eventMarshaller   EventMarshaller
}

type EventStoreConfig struct {
	Host     string
	Username string
	Password string
}

func NewEventStore(params EventStoreConfig) (EventStore, error) {
	connectionString := fmt.Sprintf("esdb://%s:%s@%s?tls=false&tlsverifycert=false", params.Username, params.Password, params.Host)
	config, err := esdb.ParseConnectionString(connectionString)

	if err != nil {
		return nil, fmt.Errorf("error parsing eventstoredb connection string: %s", err)
	}

	client, err := esdb.NewClient(config)

	if err != nil {
		return nil, fmt.Errorf("error connection to eventstoredb: %s", err)
	}

	eventTypeRegistry := make(EventTypeRegistry)

	return &eventStore{
		client:            client,
		eventTypeRegistry: eventTypeRegistry,
		eventMarshaller:   NewEventMarshaller(eventTypeRegistry),
	}, nil
}

func (es *eventStore) RegisterEventType(eventType EventType) {
	t := reflect.TypeOf(eventType).Elem()
	es.eventTypeRegistry[t.Name()] = t
}

func (es *eventStore) ReadStream(ctx context.Context, streamID string, options esdb.ReadStreamOptions, count uint64) (events []Event, err error) {
	readStream, err := es.client.ReadStream(ctx, streamID, options, count)

	if err != nil {
		return events, err
	}

	for {
		evt, _ := readStream.Recv()

		if evt == nil {
			return events, err
		}

		event, err := es.eventMarshaller.FromRecordedEvent(evt.Event)

		if err != nil {
			return events, err
		}

		events = append(events, event)
	}
}

func (es *eventStore) ReadLastEventFromStream(context context.Context, streamID string) (Event, error) {
	events, err := es.ReadStream(context, streamID, esdb.ReadStreamOptions{
		From:      esdb.End{},
		Direction: esdb.Backwards,
	}, 1)

	if err != nil {
		return nil, fmt.Errorf("error when reading stream %s: %v", streamID, err)
	}

	if len(events) > 0 {
		return events[0], nil
	}

	return nil, nil
}

func (es *eventStore) AppendToStream(ctx context.Context, streamID string, events []Event) (*esdb.WriteResult, error) {
	proposedEvents := make([]esdb.EventData, len(events))

	for i, event := range events {
		eventData, err := es.eventMarshaller.ToEventData(event)

		if err != nil {
			return nil, err
		}

		proposedEvents[i] = eventData
	}

	result, err := es.client.AppendToStream(ctx, streamID, esdb.AppendToStreamOptions{ExpectedRevision: esdb.Any{}}, proposedEvents...)

	if err != nil {
		return nil, fmt.Errorf("error when appending to stream %s: %v", streamID, err)
	}

	return result, nil
}

func (es *eventStore) SubscribeToAll(ctx context.Context, options esdb.SubscribeToAllOptions) (*esdb.Subscription, error) {
	return es.client.SubscribeToAll(ctx, options)
}

func (es *eventStore) CreatePersistentSubscription(ctx context.Context, streamName string, groupName string, options esdb.PersistentStreamSubscriptionOptions) error {
	return es.client.CreatePersistentSubscription(ctx, streamName, groupName, options)
}

func (es *eventStore) PersistentSubscribeToStream(ctx context.Context, streamName string, groupName string, options esdb.ConnectToPersistentSubscriptionOptions) (*esdb.PersistentSubscription, error) {
	return es.client.ConnectToPersistentSubscription(ctx, streamName, groupName, options)
}

func (r *eventStore) SubscribeToStream(ctx context.Context, streamID string, options esdb.SubscribeToStreamOptions) (*esdb.Subscription, error) {
	return r.client.SubscribeToStream(ctx, streamID, options)
}

func (r *eventStore) DeleteStream(ctx context.Context, streamID string) error {
	_, err := r.client.DeleteStream(ctx, streamID, esdb.DeleteStreamOptions{ExpectedRevision: esdb.Any{}})
	return err
}

func (r *eventStore) GetMarshaller() EventMarshaller {
	return r.eventMarshaller
}
