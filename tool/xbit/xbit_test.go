package xbit

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestExtract(t *testing.T) {
	input := uint32(4628440)
	a := Extract[uint32, uint16](input, 17, 15)
	b := Extract[uint32, uint16](input, 15, 2)
	c := Extract[uint32, uint16](input, 0, 15)
	fmt.Println(35, 1, 8152)
	fmt.Println(a, b, c)
}

func TestBytes(t *testing.T) {
	d := uint32(4628440)
	b := BigToBytes[uint32](d)
	var bs bytes.Buffer
	err := binary.Write(&bs, binary.BigEndian, d)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs.Bytes(), b) {
		t.Fatal()
	}
	a, err := BigFromBytes[uint32](b)
	if err != nil {
		t.Fatal(err)
	}
	var u uint32
	err = binary.Read(&bs, binary.BigEndian, &u)
	if err != nil {
		t.Fatal(err)
	}
	if d != a || a != u {
		t.Fatal()
	}

	bs.Reset()
	b = LittleToBytes[uint32](d)
	err = binary.Write(&bs, binary.LittleEndian, d)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs.Bytes(), b) {
		t.Fatal()
	}

	a, err = LittleFromBytes[uint32](b)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(&bs, binary.LittleEndian, &u)
	if err != nil {
		t.Fatal(err)
	}
	if d != a || a != u {
		t.Fatal()
	}

	d2 := int64(-4628440)
	b2 := BigToBytes[int64](d2)
	var bs2 bytes.Buffer
	err = binary.Write(&bs2, binary.BigEndian, d2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs2.Bytes(), b2) {
		t.Fatal()
	}
	a2, err := BigFromBytes[int64](b2)
	if err != nil {
		t.Fatal(err)
	}
	var u2 int64
	err = binary.Read(&bs2, binary.BigEndian, &u2)
	if err != nil {
		t.Fatal(err)
	}
	if d2 != a2 || a2 != u2 {
		t.Fatal()
	}

	bs2.Reset()
	b2 = LittleToBytes[int64](d2)
	err = binary.Write(&bs2, binary.LittleEndian, d2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bs2.Bytes(), b2) {
		t.Fatal()
	}

	a2, err = LittleFromBytes[int64](b2)
	if err != nil {
		t.Fatal(err)
	}
	err = binary.Read(&bs2, binary.LittleEndian, &u2)
	if err != nil {
		t.Fatal(err)
	}
	if d2 != a2 || a2 != u2 {
		t.Fatal()
	}
}
