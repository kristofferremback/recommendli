package sqlite

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	db  *sqlx.DB
	mux *sync.RWMutex
}

func Wrap(db *sqlx.DB) *DB {
	return &DB{
		db:  db,
		mux: &sync.RWMutex{},
	}
}

func Open(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	return db, nil
}

func (db *DB) Get(ctx context.Context) (*sqlx.DB, func()) {
	db.mux.Lock()

	return db.db, func() { db.mux.Unlock() }
}

func (db *DB) RGet(ctx context.Context) (*sqlx.DB, func()) {
	db.mux.RLock()

	return db.db, func() { db.mux.RUnlock() }
}
