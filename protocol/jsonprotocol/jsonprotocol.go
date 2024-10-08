package jsonprotocol

import (
	"encoding/json"
	"io"
)

type JsonProtocol struct{}

func (*JsonProtocol) EncodeBytes(a any) ([]byte, error) {
	return json.Marshal(a)
}

func (*JsonProtocol) DecodeBytes(b []byte, a any) error {
	return json.Unmarshal(b, a)
}

func (*JsonProtocol) Encode(w io.Writer, a any) error {
	return json.NewEncoder(w).Encode(a)
}

func (*JsonProtocol) Decode(reader io.Reader, a any) error {
	return json.NewDecoder(reader).Decode(a)
}
