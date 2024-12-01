package keyvaluestore

import (
	"context"
)

var _ KV = (*CombinedStore)(nil)

type CombinedStore struct {
	primary, secondary KV
}

func Combine(primary, secondary KV) *CombinedStore {
	return &CombinedStore{primary, secondary}
}

func (c *CombinedStore) Get(ctx context.Context, key string, out interface{}) (bool, error) {
	if exists, err := c.primary.Get(ctx, key, out); exists && err == nil {
		return true, nil
	} else if err != nil {
		return false, err
	}

	if exists, err := c.secondary.Get(ctx, key, out); exists && err == nil {
		// ensure we'll hit primary cache next time
		return exists, c.primary.Put(ctx, key, out)
	} else if err != nil {
		return false, err
	}

	return false, nil
}

func (c *CombinedStore) Put(ctx context.Context, key string, data interface{}) error {
	if err := c.primary.Put(ctx, key, data); err != nil {
		return err
	}
	return c.secondary.Put(ctx, key, data)
}
