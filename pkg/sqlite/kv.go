package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
	"github.com/zmb3/spotify"
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

func (kv *KV) Get(ctx context.Context, key string, out any) (bool, error) {
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

func (kv *KV) GetMany(ctx context.Context, keys []string, out any) error {
	if len(keys) == 0 {
		return nil
	}

	db, unlock := kv.db.RGet(ctx)
	defer unlock()

	query, args, err := sqlx.In(`
		SELECT key, value
		FROM keyvaluestore
		WHERE kind = ?
			AND key IN (?)
	`, kv.kind, keys)
	if err != nil {
		return fmt.Errorf("building query: %w", err)
	}

	rows, err := db.QueryContext(ctx, sqlx.Rebind(sqlx.QUESTION, query), args...)
	if err != nil {
		return fmt.Errorf("getting many from keyvaluestore: %w", err)
	}
	defer rows.Close()

	// Create a sparse map of the values from the database so we can get the values out in order.
	values := make(map[string][]byte)
	for i := 0; rows.Next(); i++ {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("scanning keyvaluestore: %w", err)
		}
		values[key] = value
	}

	// Build up an array of all items in the same order as keys,
	// filling in the blanks with null values.
	allValues := []byte(`[`)
	for i, key := range keys {
		if value, ok := values[key]; ok {
			allValues = append(allValues, value...)
		} else {
			allValues = append(allValues, []byte(`null`)...) // Shitty way of doing this, assuming null is a valid value...
		}
		if i < len(keys)-1 {
			allValues = append(allValues, []byte(`,`)...)
		}
	}
	allValues = append(allValues, []byte(`]`)...)

	if err := kv.unmarshalValue(allValues, out); err != nil {
		return fmt.Errorf("unmarshalling values: %w", err)
	}

	return nil
}

func (kv *KV) Put(ctx context.Context, key string, data any) error {
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

func (kv *KV) marshalValue(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (kv *KV) unmarshalValue(data []byte, out any) error {
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("unmarshalling value: %w", err)
	}

	switch res := out.(type) {
	case *spotify.FullPlaylist:
		// The spotify.FullPlaylist wraps a spotify.SimplePlaylist except
		// the tracks, which results in the simple playlist missing track
		// information.
		res.SimplePlaylist.Tracks = spotify.PlaylistTracks{
			Endpoint: res.Tracks.Endpoint,
			Total:    uint(res.Tracks.Total),
		}
	case *[]spotify.FullPlaylist:
		for i := range *res {
			pl := (*res)[i]
			pl.SimplePlaylist.Tracks = spotify.PlaylistTracks{
				Endpoint: pl.Tracks.Endpoint,
				Total:    uint(pl.Tracks.Total),
			}
			(*res)[i] = pl
		}
	}

	return nil
}
