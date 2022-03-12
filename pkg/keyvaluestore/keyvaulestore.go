package keyvaluestore

import (
	"context"
	"io"
)

type Serializer interface {
	Serialize(writer io.Writer, data interface{}) error
	Deserialize(reader io.Reader, out interface{}) error
}

type KV interface {
	Get(ctx context.Context, key string, out interface{}) (bool, error)
	Put(ctx context.Context, key string, data interface{}) error
	List(ctx context.Context) ([]string, error)
}
