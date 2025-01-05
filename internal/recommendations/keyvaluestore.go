package recommendations

import "context"

type KeyValueStore interface {
	Get(ctx context.Context, key string, out any) (bool, error)
	GetMany(ctx context.Context, keys []string, out any) error
	Put(ctx context.Context, key string, data any) error
}
