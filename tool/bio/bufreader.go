package bio

import "bytes"

type readerPacket interface {
	ReadPacket() ([]byte, error)
}

type BufferReader struct {
	min int
	max int

	buf bytes.Buffer
	rp  readerPacket
}

func NewBufferReader(rp readerPacket) *BufferReader {
	return &BufferReader{
		min: 0,
		max: 0,
		buf: bytes.Buffer{},
		rp:  rp,
	}
}

func NewBufferReaderWithOpt(rp readerPacket, min, max int) *BufferReader {
	if min < 0 {
		min = 0
	}
	if max < 0 {
		max = 0
	}
	var buf bytes.Buffer
	buf.Grow(max)
	return &BufferReader{
		min: min,
		max: max,
		buf: buf,
		rp:  rp,
	}
}

func (br *BufferReader) Read(b []byte) (n int, err error) {
	for br.buf.Len() <= br.min {
		err = br.toBuffer()
		if err != nil {
			return 0, err
		}
	}
	return br.buf.Read(b)
}

func (br *BufferReader) toBuffer() error {
	packet, err := br.rp.ReadPacket()
	if err != nil {
		return err
	}
	br.buf.Write(packet)
	return nil
}
