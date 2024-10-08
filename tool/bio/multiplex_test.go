package bio

import (
	"bytes"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"io"
	"testing"
)

func TestMultiplexIO(t *testing.T) {
	reader, writer := io.Pipe()
	r := NewRWC(reader, writer, nil)
	multiplexIO := NewMultiplexIO(r)
	rid := uint32(1230)
	lid := uint32(1231)
	b := []byte(uuid.NewIdn(4096))
	go func() {
		err := multiplexIO.WriteMsg(b, rid, lid)
		if err != nil {
			t.Error(err)
		}
	}()
	msg, rrid, rlid, err := multiplexIO.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}
	if rrid != rid || rlid != lid || !bytes.Equal(msg, b) {
		t.Fatal()
	}
}

func TestMultiplexIOSize(t *testing.T) {
	reader, writer := io.Pipe()
	r := NewRWC(reader, writer, nil)
	multiplexIO := NewMultiplexIO(r)
	rid := uint32(1230)
	lid := uint32(1231)
	b := []byte(uuid.NewIdn(4096))
	go func() {
		err := multiplexIO.WriteMsgWithSize(b, rid, lid, 100)
		if err != nil {
			t.Error(err)
		}
	}()
	l := len(b)
	s := 0
	var bs [][]byte
	for l != s {
		msg, rrid, rlid, err := multiplexIO.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		if rrid != rid || rlid != lid {
			t.Fatal()
		}
		bs = append(bs, msg)
		s += len(msg)
	}

	if !bytes.Equal(bytes.Join(bs, nil), b) {
		t.Fatal()
	}
}

func TestMultiplexIOPreSize(t *testing.T) {
	reader, writer := io.Pipe()
	r := NewRWC(reader, writer, nil)
	multiplexIO := NewMultiplexIO(r)
	rid := uint32(1230)
	lid := uint32(1231)
	b := []byte(uuid.NewIdn(4096))
	pre := []byte{1, 2, 3}
	go func() {
		err := multiplexIO.WriteMsgWithPreSize(b, pre, rid, lid, 100)
		if err != nil {
			t.Error(err)
		}
	}()
	l := len(b)
	s := 0
	var bs [][]byte
	for l != s {
		msg, rrid, rlid, err := multiplexIO.ReadMsg()
		if err != nil {
			t.Fatal(err)
		}
		if rrid != rid || rlid != lid {
			t.Fatal()
		}
		msg, _ = bytes.CutPrefix(msg, pre)
		bs = append(bs, msg)
		s += len(msg)
	}

	if !bytes.Equal(bytes.Join(bs, nil), b) {
		t.Fatal()
	}
}
