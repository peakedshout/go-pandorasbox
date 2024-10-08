package protocol

import (
	"io"
)

type Protocol interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(w io.Writer, a any) error
}

type Decoder interface {
	Decode(r io.Reader, a any) error
}

type BytesProtocol interface {
	BytesEncoder
	BytesDecoder
}

type BytesEncoder interface {
	EncodeBytes(a any) ([]byte, error)
}

type BytesDecoder interface {
	DecodeBytes(b []byte, a any) error
}
