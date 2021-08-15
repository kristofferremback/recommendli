package keyvaluestore

import (
	"errors"
	"io"
)

var ErrNoSuchItem = errors.New("no such item")

type Serializer interface {
	Serialize(writer io.Writer, data interface{}) error
	Deserialize(reader io.Reader, out interface{}) error
}
