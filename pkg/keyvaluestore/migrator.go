package keyvaluestore

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/logging"
)

type Migrator struct {
	log            logging.Logger
	source, target KV
}

func NewMigrator(log logging.Logger, source, target KV) *Migrator {
	return &Migrator{log, source, target}
}

func (m *Migrator) Migrate(ctx context.Context) error {
	keys, err := m.source.List(ctx)
	if err != nil {
		return fmt.Errorf("listing keys from source: %w", err)
	}
	m.log.Info("listed keys from source", "keyCount", len(keys))

	for _, key := range keys {
		m.log.Info("copying key from source to target", "key", key)
		var value interface{}
		if _, err := m.source.Get(ctx, key, &value); err != nil {
			return fmt.Errorf("getting value of %s from source: %w", key, err)
		}
		if err := m.target.Put(ctx, key, value); err != nil {
			return fmt.Errorf("putting value of %s in target: %w", key, err)
		}
	}
	m.log.Info("successfully migrated keys from source to target", "keyCount", len(keys))
	return nil
}
