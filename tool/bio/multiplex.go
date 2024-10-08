package bio

import (
	"bytes"
	"errors"
	"github.com/peakedshout/go-pandorasbox/tool/xbit"
	"io"
)

const (
	multiplexHeaderSize = 12
	multiplexRIdSize    = 4
	multiplexLIdSize    = 4
	multiplexLenSize    = 4
)

func NewMultiplexIO(rwc io.ReadWriteCloser) *MultiplexIO {
	return &MultiplexIO{rwc: rwc}
}

type MultiplexIO struct {
	rwc io.ReadWriteCloser
}

func (mio *MultiplexIO) ReadMsg() ([]byte, uint32, uint32, error) {
	buf := make([]byte, multiplexHeaderSize)
	_, err := io.ReadFull(mio.rwc, buf)
	if err != nil {
		return nil, 0, 0, err
	}
	rid, _ := xbit.BigFromBytes[uint32](buf[0:multiplexRIdSize])
	lid, _ := xbit.BigFromBytes[uint32](buf[multiplexRIdSize : multiplexRIdSize+multiplexLIdSize])
	lens, _ := xbit.BigFromBytes[uint32](buf[multiplexRIdSize+multiplexLIdSize : multiplexRIdSize+multiplexLIdSize+multiplexLenSize])
	buf = make([]byte, lens)
	_, err = io.ReadFull(mio.rwc, buf)
	if err != nil {
		return nil, 0, 0, err
	}
	return buf, rid, lid, nil
}

func (mio *MultiplexIO) WriteMsg(b []byte, rid, lid uint32) error {
	ridbs := xbit.BigToBytes[uint32](rid)
	lidbs := xbit.BigToBytes[uint32](lid)
	lens := xbit.BigToBytes[uint32](uint32(len(b)))
	_, err := mio.rwc.Write(bytes.Join([][]byte{ridbs, lidbs, lens, b}, nil))
	return err
}

func (mio *MultiplexIO) WriteMsgWithSize(b []byte, rid, lid uint32, size int) error {
	if size == 0 {
		return mio.WriteMsg(b, rid, lid)
	}
	ridbs := xbit.BigToBytes[uint32](rid)
	lidbs := xbit.BigToBytes[uint32](lid)
	lens := len(b)
	for j := 0; j < lens; j += size {
		next := j + size
		if lens <= next {
			next = lens
		}
		lenbs := xbit.BigToBytes[uint32](uint32(len(b[j:next])))
		_, err := mio.rwc.Write(bytes.Join([][]byte{ridbs, lidbs, lenbs, b[j:next]}, nil))
		if err != nil {
			return err
		}
	}
	return nil
}

func (mio *MultiplexIO) WriteMsgWithPreSize(b, pre []byte, rid, lid uint32, size int) error {
	if size == 0 {
		return mio.WriteMsg(bytes.Join([][]byte{pre, b}, nil), rid, lid)
	}
	size -= len(pre)
	if size <= 0 {
		return errors.New("pre size too long")
	}
	ridbs := xbit.BigToBytes[uint32](rid)
	lidbs := xbit.BigToBytes[uint32](lid)
	lens := len(b)
	for j := 0; j < lens; j += size {
		next := j + size
		if lens <= next {
			next = lens
		}
		lenbs := xbit.BigToBytes[uint32](uint32(len(b[j:next]) + len(pre)))
		_, err := mio.rwc.Write(bytes.Join([][]byte{ridbs, lidbs, lenbs, pre, b[j:next]}, nil))
		if err != nil {
			return err
		}
	}
	return nil
}

func (mio *MultiplexIO) Close() error {
	return mio.rwc.Close()
}
