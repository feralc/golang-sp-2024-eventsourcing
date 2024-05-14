package projection

import (
	"context"
	"database/sql"
	"log"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/feralc/golang-sp-2024-eventsourcing/application/service"
	"github.com/feralc/golang-sp-2024-eventsourcing/esourcing"
)

type EventProjectionHandleFunc func(ctx context.Context, evt esourcing.Event, tx *sql.Tx) (err error)

type Projection struct {
	svc                 *service.Service
	store               esourcing.EventStore
	subscriptionManager esourcing.SubscriptionManager
	projectionName      string
	streamName          string
	groupName           string
	isPersistent        bool
}

func NewProjection(svc *service.Service, store esourcing.EventStore, projectionName string) *Projection {
	return &Projection{
		svc:                 svc,
		store:               store,
		subscriptionManager: esourcing.NewSubscriptionManager(svc.GetBD()),
		projectionName:      projectionName,
		isPersistent:        false,
	}
}

func NewProjectionWithPersistentSubscription(svc *service.Service, store esourcing.EventStore, projectionName string, streamName string, groupName string) *Projection {
	return &Projection{
		svc:                 svc,
		store:               store,
		projectionName:      projectionName,
		streamName:          streamName,
		groupName:           groupName,
		isPersistent:        true,
		subscriptionManager: esourcing.NewSubscriptionManager(svc.GetBD()),
	}
}

func (p *Projection) Run(ctx context.Context, evtPrefixes []string, handleEventFunc EventProjectionHandleFunc) {
	if err := p.subscriptionManager.CreateSubscriptionIfNotExists(p.projectionName); err != nil {
		panic(err)
	}

	startFrom, err := p.getStartFrom(ctx)
	if err != nil {
		panic(err)
	}

	if p.isPersistent {
		subscription, err := p.getPersistentSubscription(ctx, esdb.ConnectToPersistentSubscriptionOptions{})
		if err != nil {
			log.Fatalln(err.Error())
			panic(err)
		}

		err = p.handleEventsFromPersistentSubscription(ctx, subscription, handleEventFunc)
		if err != nil {
			panic(err)
		}
		return
	}

	subscriptionOptions := esdb.SubscribeToAllOptions{
		From: startFrom,
		Filter: &esdb.SubscriptionFilter{
			Type:     esdb.EventFilterType,
			Prefixes: evtPrefixes,
		},
	}

	subscription, err := p.getSubscription(ctx, subscriptionOptions)
	if err != nil {
		log.Fatalln(err.Error())
		panic(err)
	}

	err = p.handleEventsFromSubscription(ctx, subscription, handleEventFunc)
	if err != nil {
		panic(err)
	}
}

func (p *Projection) getSubscription(ctx context.Context, subscriptionOptions esdb.SubscribeToAllOptions) (subscription *esdb.Subscription, err error) {
	subscription, err = p.store.SubscribeToAll(ctx, subscriptionOptions)
	if err != nil {
		return subscription, err
	}

	return subscription, nil
}

func (p *Projection) getPersistentSubscription(ctx context.Context, subscriptionOptions esdb.ConnectToPersistentSubscriptionOptions) (subscription *esdb.PersistentSubscription, err error) {
	subscription, err = p.store.PersistentSubscribeToStream(ctx, p.streamName, p.groupName, subscriptionOptions)
	if err != nil {
		return subscription, err
	}

	return subscription, nil
}

func (p *Projection) getStartFrom(ctx context.Context) (startFrom esdb.AllPosition, err error) {
	startFrom = esdb.Start{}
	lastCheckpointPosition, err := p.subscriptionManager.LastCheckpoint(p.projectionName)
	if err != nil {
		return startFrom, err
	}

	if lastCheckpointPosition != nil {
		startFrom = *lastCheckpointPosition
	}

	return startFrom, nil
}

func (p *Projection) handleEventsFromSubscription(ctx context.Context, subscription *esdb.Subscription, handleEventFunc EventProjectionHandleFunc) (err error) {
	defer subscription.Close()

	for {
		evt := subscription.Recv()

		tx, err := p.svc.GetBD().BeginTx(ctx, nil)
		if err != nil {
			panic(err)
		}

		defer tx.Rollback()

		if evt.EventAppeared != nil {
			event, err := p.store.GetMarshaller().FromRecordedEvent(evt.EventAppeared.Event)
			if err != nil {
				log.Println(err)
			}

			err = handleEventFunc(ctx, event, tx)
			if err != nil {
				return err
			}

			if evt.CheckPointReached == nil {
				err = p.subscriptionManager.SaveCheckpoint(p.projectionName, &evt.EventAppeared.OriginalEvent().Position, tx)
				if err != nil {
					panic(err)
				}
			}

			log.Printf("Processed event %s@%s\n", evt.EventAppeared.Event.EventType, evt.EventAppeared.Event.EventID)
		}

		if evt.CheckPointReached != nil {
			err = p.subscriptionManager.SaveCheckpoint(p.projectionName, evt.CheckPointReached, tx)
			if err != nil {
				panic(err)
			}
		}

		if evt.SubscriptionDropped != nil {
			log.Printf("subscription dropped: %v", evt.SubscriptionDropped.Error)
			break
		}

		tx.Commit()
	}

	return nil
}

func (p *Projection) handleEventsFromPersistentSubscription(ctx context.Context, subscription *esdb.PersistentSubscription, handleEventFunc EventProjectionHandleFunc) (err error) {
	defer subscription.Close()

	for {
		evt := subscription.Recv()

		tx, err := p.svc.GetBD().BeginTx(ctx, nil)
		if err != nil {
			panic(err)
		}

		defer tx.Rollback()

		if evt.EventAppeared != nil {
			event, err := p.store.GetMarshaller().FromRecordedEvent(evt.EventAppeared.Event)
			if err != nil {
				log.Println(err)
			}

			err = handleEventFunc(ctx, event, tx)
			if err != nil {
				return err
			}

			log.Printf("Processed event %s@%s\n", evt.EventAppeared.Event.EventType, evt.EventAppeared.Event.EventID)
		}

		if evt.CheckPointReached != nil {
			err = p.subscriptionManager.SaveCheckpoint(projectionName, evt.CheckPointReached, tx)
			if err != nil {
				panic(err)
			}
		}

		if evt.SubscriptionDropped != nil {
			log.Printf("subscription dropped: %v", evt.SubscriptionDropped.Error)
			break
		}

		tx.Commit()
	}

	return nil
}
