package service

import (
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) (*Service, error) {
	svc := new(Service)
	svc.db = db

	return svc, nil
}

func (svc *Service) GetBD() *sql.DB {
	return svc.db
}
