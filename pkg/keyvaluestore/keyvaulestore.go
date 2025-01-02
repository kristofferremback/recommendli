package keyvaluestore

import (
	"context"
	"io"
)

type Serializer interface {
	Serialize(writer io.Writer, data any) error
	Deserialize(reader io.Reader, out any) error
}

type KV interface {
	Get(ctx context.Context, key string, out any) (bool, error)
	GetMany(ctx context.Context, keys []string, out any) error
	Put(ctx context.Context, key string, data any) error
}
