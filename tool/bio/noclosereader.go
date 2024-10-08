package bio

import (
	"bytes"
	"io"
)

type noCloseReader struct {
	*bytes.Reader
}

func NewNoCloseBody(b []byte) io.ReadCloser {
	return &noCloseReader{
		Reader: bytes.NewReader(b),
	}
}

func (ncb *noCloseReader) Close() error {
	return nil
}
