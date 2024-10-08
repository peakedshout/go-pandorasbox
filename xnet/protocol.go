package xnet

import (
	"encoding/json"
	"io"
)

type Protocol interface {

	// Encode
	//  The cutting strategy of the data should be controlled by the upper layer,
	//  and the protocol can only do simple cuts
	Encode(a any) ([]byte, error)

	// Decode
	//  The internal implementation should try to parse the content, and if it cannot be parsed,
	//  rollback can be considered, but it is not necessary
	Decode(reader io.Reader, a any) error
}

type NoProtocol struct {
}

func (*NoProtocol) Encode(a any) ([]byte, error) {
	b, err := json.Marshal(a)
	return b, err
}

func (*NoProtocol) Decode(reader io.Reader, a any) error {
	return json.NewDecoder(reader).Decode(a)
}
