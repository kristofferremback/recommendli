package keyvaluestore

import (
	"encoding/json"
	"io"
)

type JSONSerializer struct{}

func (s JSONSerializer) Serialize(writer io.Writer, data interface{}) error {
	return json.NewEncoder(writer).Encode(data)
}

func (s JSONSerializer) Deserialize(reader io.Reader, out interface{}) error {
	return json.NewDecoder(reader).Decode(out)
}
