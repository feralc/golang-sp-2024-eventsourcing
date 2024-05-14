package esourcing

import (
	"time"

	"github.com/gofrs/uuid"
)

type EventBase struct {
	eventID       string
	eventType     string
	timestamp     time.Time
	aggregateType AggregateType
	aggregateID   string
}

func NewEventBaseForAggregate(aggregateType AggregateType, aggregateID string, eventType string, timestamp time.Time) *EventBase {
	eventID, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	return &EventBase{
		eventID:       eventID.String(),
		eventType:     eventType,
		timestamp:     timestamp,
		aggregateType: aggregateType,
		aggregateID:   aggregateID,
	}
}

func NewEventBaseForAggregateWithID(eventID string, aggregateType AggregateType, aggregateID string, eventType string, timestamp time.Time) *EventBase {
	event := NewEventBaseForAggregate(aggregateType, aggregateID, eventType, timestamp)
	event.eventID = eventID
	return event
}

func (e *EventBase) EventID() string {
	return e.eventID
}

func (e *EventBase) EventType() string {
	return e.eventType
}

func (e *EventBase) Timestamp() time.Time {
	return e.timestamp
}

func (e *EventBase) AggregateType() AggregateType {
	return e.aggregateType
}

func (e *EventBase) AggregateID() string {
	return e.aggregateID
}
