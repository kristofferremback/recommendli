package keyvaluestore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/replit/database-go"
)

var _ KV = (*ReplitStore)(nil)

type ReplitStore struct {
	prefix     string
	serializer Serializer
}

func ReplitDBJSONStore(prefix string) *ReplitStore {
	return &ReplitStore{prefix: prefix, serializer: &JSONSerializer{}}
}

func (d *ReplitStore) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	s, err := database.Get(d.keyOf(key))
	if err != nil && errors.Is(err, database.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("getting value for %s: %w", d.keyOf(key), err)
	}
	if err := d.serializer.Deserialize(strings.NewReader(s), &out); err != nil {
		return false, fmt.Errorf("deserializing data: %w", err)
	}
	return true, nil
}

func (d *ReplitStore) Put(ctx context.Context, key string, data interface{}) error {
	writer := strings.Builder{}
	if err := d.serializer.Serialize(&writer, data); err != nil {
		return fmt.Errorf("deserializing data: %w", err)
	}
	if err := database.Set(d.keyOf(key), writer.String()); err != nil {
		return fmt.Errorf("setting value for %s: %w", d.keyOf(key), err)
	}
	return nil
}

func (d *ReplitStore) List(ctx context.Context) ([]string, error) {
	keys, err := database.ListKeys(d.prefix)
	if err != nil {
		return nil, fmt.Errorf("listing keys for %s: %w", d.prefix, err)
	}
	return keys, nil
}

func (d *ReplitStore) keyOf(key string) string {
	return fmt.Sprintf("%s/%s", d.prefix, key)
}
