package esourcing

import (
	"reflect"
	"sync"
	"time"
)

type AggregateType string

func (at AggregateType) String() string {
	return string(at)
}

type Aggregate interface {
	AggregateID() string
	AggregateType() AggregateType
	Events() []Event
	SetEvents(events []Event)
	ApplyEvent(event Event)
	UncommittedEvents() []Event
	SetUncommittedEvents(events []Event)
	ClearUncommittedEvents()
	GetAndClearUncommitedEvents() []Event
}

type AggregateRoot struct {
	id                string
	aggregateType     AggregateType
	events            []Event
	uncommittedEvents []Event
	mu                *sync.Mutex
}

func NewAggregateRoot(aggregateType AggregateType, aggregateID string) *AggregateRoot {
	return &AggregateRoot{
		id:            aggregateID,
		aggregateType: aggregateType,
		mu:            &sync.Mutex{},
	}
}

func (a *AggregateRoot) AggregateID() string {
	return a.id
}

func (a *AggregateRoot) AggregateType() AggregateType {
	return a.aggregateType
}

func (a *AggregateRoot) Events() []Event {
	return a.events
}

func (a *AggregateRoot) SetEvents(events []Event) {
	a.events = events
}

func (a *AggregateRoot) SetUncommittedEvents(events []Event) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.uncommittedEvents = events
}

func (a *AggregateRoot) UncommittedEvents() []Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.uncommittedEvents
}

func (a *AggregateRoot) ClearUncommittedEvents() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.uncommittedEvents = nil
}

func (a *AggregateRoot) GetAndClearUncommitedEvents() []Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	events := a.uncommittedEvents
	a.uncommittedEvents = nil
	return events
}

func AppendEvent(agg Aggregate, event Event) {
	// v is the interface{}
	v := reflect.ValueOf(&event).Elem()

	// Allocate a temporary variable with type of the struct
	// v.Elem() is the vale contained in the interface
	tmp := reflect.New(v.Elem().Type()).Elem()

	// Copy the struct value contained in interface to
	// the temporary variable
	tmp.Set(v.Elem())

	eventType := tmp.Type().Name()
	baseEvent := NewEventBaseForAggregate(agg.AggregateType(), agg.AggregateID(), eventType, time.Now())
	tmp.FieldByName("EventBase").Set(reflect.ValueOf(baseEvent))

	// Set the interface to the modified struct value
	v.Set(tmp)

	agg.SetUncommittedEvents(append(agg.UncommittedEvents(), event))
	agg.ApplyEvent(event)
}

func Commit(a Aggregate) {
	a.SetEvents(append(a.Events(), a.UncommittedEvents()...))
	a.ClearUncommittedEvents()
}

func RebuildFromEvents(a Aggregate, events []Event) {
	a.SetEvents(events)

	for _, e := range events {
		a.ApplyEvent(e)
	}
}
