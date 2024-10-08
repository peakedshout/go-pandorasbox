package gobprotocol

import (
	"bytes"
	"encoding/gob"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"io"
)

type GobProtocol struct {
	crypto pcrypto.PCrypto
}

func (g *GobProtocol) Encode(w io.Writer, a any) error {
	return gob.NewEncoder(w).Encode(a)
}

func (g *GobProtocol) Decode(r io.Reader, a any) error {
	return gob.NewDecoder(r).Decode(a)
}

func (g *GobProtocol) EncodeBytes(a any) ([]byte, error) {
	var bs bytes.Buffer
	err := g.Encode(&bs, a)
	if err != nil {
		return nil, err
	}
	return bs.Bytes(), nil
}

func (g *GobProtocol) DecodeBytes(b []byte, a any) error {
	r := bytes.NewReader(b)
	return g.Decode(r, a)
}
