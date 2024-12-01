package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
)

var _ keyvaluestore.KV = (*KV)(nil)

type KV struct {
	db   *DB
	kind string
}

func NewKV(db *DB, kind string) *KV {
	return &KV{
		db:   db,
		kind: kind,
	}
}

func (kv *KV) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	db, unlock := kv.db.RGet(ctx)
	defer unlock()

	row := db.QueryRowContext(ctx, `SELECT value FROM keyvaluestore WHERE kind = ? AND key = ?`, kv.kind, key)
	if err := row.Err(); err != nil {
		return false, fmt.Errorf("querying keyvaluestore: %w", err)
	}

	var value []byte
	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("scanning keyvaluestore: %w", err)
	}

	if err := kv.unmarshalValue(value, out); err != nil {
		return false, fmt.Errorf("unmarshalling value: %w", err)
	}

	return true, nil
}

func (kv *KV) Put(ctx context.Context, key string, data interface{}) error {
	db, unlock := kv.db.Get(ctx)
	defer unlock()

	value, err := kv.marshalValue(data)
	if err != nil {
		return fmt.Errorf("marshalling value: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT OR REPLACE INTO keyvaluestore (key, kind, value, updated_at)
		VALUES (?, ?, ?, datetime('now'))
	`, key, kv.kind, value); err != nil {
		return fmt.Errorf("inserting into keyvaluestore: %w", err)
	}

	return nil
}

func (kv *KV) marshalValue(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (kv *KV) unmarshalValue(data []byte, out interface{}) error {
	return json.Unmarshal(data, out)
}
