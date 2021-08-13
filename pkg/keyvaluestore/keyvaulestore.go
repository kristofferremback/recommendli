package keyvaluestore

import (
	"io"
)

type Serializer interface {
	Serialize(writer io.Writer, data interface{}) error
	Deserialize(reader io.Reader, out interface{}) error
}
