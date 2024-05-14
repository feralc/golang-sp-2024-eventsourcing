package esourcing

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/esdb"
)

type subscriptionManager struct {
	db *sql.DB
}

func NewSubscriptionManager(db *sql.DB) SubscriptionManager {
	return &subscriptionManager{
		db: db,
	}
}

func (sm *subscriptionManager) CreateSubscriptionIfNotExists(subscriptionID string) error {
	_, err := sm.db.Exec("insert ignore into es_subscription_checkpoint (subscription_id) values (?);", subscriptionID)
	return err
}

func (sm *subscriptionManager) LastCheckpoint(subscriptionID string) (*esdb.Position, error) {
	row := sm.db.QueryRow("SELECT coalesce(checkpoint_position, '') FROM es_subscription_checkpoint WHERE subscription_id = ?", subscriptionID)

	var lastPosition string

	err := row.Scan(&lastPosition)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if lastPosition == "" {
		return nil, nil
	}

	positions := strings.Split(lastPosition, ":")
	prepare, _ := strconv.ParseUint(positions[0], 10, 64)
	commit, _ := strconv.ParseUint(positions[1], 10, 64)

	return &esdb.Position{
		Prepare: prepare,
		Commit:  commit,
	}, nil
}

func (sm *subscriptionManager) SaveCheckpoint(subscriptionID string, position *esdb.Position, tx *sql.Tx) error {
	checkpointPosition := fmt.Sprintf("%v:%v", position.Prepare, position.Commit)
	_, err := tx.Exec("UPDATE es_subscription_checkpoint SET checkpoint_position = ?, checkpoint_at = now() WHERE subscription_id = ?", checkpointPosition, subscriptionID)
	return err
}
