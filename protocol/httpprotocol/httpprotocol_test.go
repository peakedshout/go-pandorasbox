package httpprotocol

import (
	"bytes"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func TestNewHttpProtocol(t *testing.T) {
	chp := NewHttpProtocol(true, nil, nil, nil, nil, nil)
	shp := NewHttpProtocol(false, nil, nil, nil, nil, nil)
	bs := new(bytes.Buffer)
	data := uuid.NewId(1)
	err := chp.Encode(bs, data)
	if err != nil {
		t.Fatal(err)
	}
	var r string
	err = shp.Decode(bs, &r)
	if err != nil {
		t.Fatal(err)
	}
	if data != r {
		t.Fatal()
	}
	data = uuid.NewId(1)
	err = shp.Encode(bs, data)
	if err != nil {
		t.Fatal(err)
	}
	err = chp.Decode(bs, &r)
	if err != nil {
		t.Fatal(err)
	}
	if data != r {
		t.Fatal()
	}
}
