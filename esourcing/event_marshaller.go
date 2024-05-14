package esourcing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/gofrs/uuid"
)

type eventMarshaller struct {
	eventTypeRegistry EventTypeRegistry
}

func NewEventMarshaller(eventTypeRegistry EventTypeRegistry) EventMarshaller {
	return &eventMarshaller{
		eventTypeRegistry,
	}
}

func (em *eventMarshaller) ToEventData(event Event) (esdb.EventData, error) {
	eventData, err := json.Marshal(event)

	if err != nil {
		return esdb.EventData{}, fmt.Errorf("error when marshalling event %v: %v", event.EventType(), err)
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"version":        event.Version(),
		"aggregate_type": event.AggregateType(),
		"aggregate_id":   event.AggregateID(),
		"timestamp":      event.Timestamp(),
	})

	if err != nil {
		return esdb.EventData{}, fmt.Errorf("error when marshalling event %v: %v", event.EventType(), err)
	}

	eventID, err := uuid.FromString(event.EventID())

	if err != nil {
		return esdb.EventData{}, fmt.Errorf("error when marshalling event %v: %v", event.EventType(), err)
	}

	return esdb.EventData{
		EventID:     eventID,
		EventType:   event.EventType(),
		ContentType: esdb.JsonContentType,
		Data:        eventData,
		Metadata:    metadata,
	}, nil
}

func (em *eventMarshaller) FromRecordedEvent(recordedEvent *esdb.RecordedEvent) (Event, error) {
	eventReflectType, ok := em.eventTypeRegistry[recordedEvent.EventType]

	if !ok {
		return nil, fmt.Errorf("unknown event type %s", recordedEvent.EventType)
	}

	eventReflect := reflect.New(eventReflectType)
	eventReflectElm := eventReflect.Elem()
	metadata := map[string]string{}

	err := json.Unmarshal(recordedEvent.UserMetadata, &metadata)

	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling event %v: %v", recordedEvent.EventType, err)
	}

	var eventData map[string]interface{}
	err = json.Unmarshal(recordedEvent.Data, &eventData)

	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling event %v: %v", recordedEvent.EventType, err)
	}

	eventDataBytes, err := json.Marshal(eventData)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(eventDataBytes, eventReflect.Interface())

	if err != nil {
		return nil, err
	}

	timestamp, err := time.Parse(time.RFC3339, metadata["timestamp"])

	if err != nil {
		return nil, err
	}

	eventBase := NewEventBaseForAggregateWithID(
		recordedEvent.EventID.String(),
		AggregateType(metadata["aggregate_type"]),
		metadata["aggregate_id"],
		recordedEvent.EventType,
		timestamp,
	)

	eventReflectElm.FieldByName("EventBase").Set(reflect.ValueOf(eventBase))
	domainEvent, _ := eventReflectElm.Interface().(Event)

	return domainEvent, nil
}
