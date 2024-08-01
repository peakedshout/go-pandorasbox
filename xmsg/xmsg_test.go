package xmsg

import (
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func TestXMsg(t *testing.T) {
	data := uuid.NewId(3)
	xMsg1, err := newXMsg("test1", 1, 21232, 23, data)
	if err != nil {
		t.Fatal(err)
	}
	b, err := xMsg1.marshal()
	if err != nil {
		t.Fatal(err)
	}
	xMsg2 := &XMsg{}
	err = xMsg2.unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if xMsg1.header != xMsg2.header || xMsg1.id != xMsg2.id || xMsg1.flag != xMsg2.flag || xMsg1.opt != xMsg2.opt {
		fmt.Println(xMsg1.id, xMsg2.id)
		t.Fatal()
	}
	str := ""
	err = xMsg2.Unmarshal(&str)
	if err != nil {
		t.Fatal(err)
	}
	if str != data {
		t.Fatal()
	}
}

func TestXMsgError(t *testing.T) {
	data := errors.New("test error")
	xMsg1, err := newXMsg("test1", 1, 21232, 23, data)
	if err != nil {
		t.Fatal(err)
	}
	b, err := xMsg1.marshal()
	if err != nil {
		t.Fatal(err)
	}
	xMsg2 := &XMsg{}
	err = xMsg2.unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if xMsg1.header != xMsg2.header || xMsg1.id != xMsg2.id || xMsg1.flag != xMsg2.flag || xMsg1.opt != xMsg2.opt {
		t.Fatal()
	}
	err = xMsg2.Unmarshal(nil)
	if err == nil {
		t.Fatal()
	}
	if err.Error() != data.Error() {
		t.Fatal()
	}
}

func TestXMsgStruct(t *testing.T) {
	type xxx struct {
		s string
		I int
		S string
	}
	data := xxx{
		s: "a",
		I: 222,
		S: "bbb",
	}
	xMsg1, err := newXMsg("test1", 1, 21232, 23, data)
	if err != nil {
		t.Fatal(err)
	}
	b, err := xMsg1.marshal()
	if err != nil {
		t.Fatal(err)
	}
	xMsg2 := &XMsg{}
	err = xMsg2.unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if xMsg1.header != xMsg2.header || xMsg1.id != xMsg2.id || xMsg1.flag != xMsg2.flag || xMsg1.opt != xMsg2.opt {
		t.Fatal()
	}
	var x xxx
	err = xMsg2.Unmarshal(&x)
	if err != nil {
		t.Fatal(err)
	}
	if x.S != data.S || x.I != data.I {
		t.Fatal()
	}
}

func TestXMsgPtr(t *testing.T) {
	type xxx struct {
		s string
		I int
		S string
	}
	data := &xxx{
		s: "a",
		I: 222,
		S: "bbb",
	}
	xMsg1, err := newXMsg("test1", 1, 21232, 23, data)
	if err != nil {
		t.Fatal(err)
	}
	b, err := xMsg1.marshal()
	if err != nil {
		t.Fatal(err)
	}
	xMsg2 := &XMsg{}
	err = xMsg2.unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if xMsg1.header != xMsg2.header || xMsg1.id != xMsg2.id || xMsg1.flag != xMsg2.flag || xMsg1.opt != xMsg2.opt {
		t.Fatal()
	}
	var x *xxx
	err = xMsg2.Unmarshal(&x)
	if err != nil {
		t.Fatal(err)
	}
	if x.S != data.S || x.I != data.I {
		t.Fatal()
	}
}

func TestXMsgPtr2(t *testing.T) {
	type xxx struct {
		s string
		I int
		S string
	}
	var data *xxx
	xMsg1, err := newXMsg("test1", 1, 21232, 23, data)
	if err != nil {
		t.Fatal(err)
	}
	b, err := xMsg1.marshal()
	if err != nil {
		t.Fatal(err)
	}
	xMsg2 := &XMsg{}
	err = xMsg2.unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if xMsg1.header != xMsg2.header || xMsg1.id != xMsg2.id || xMsg1.flag != xMsg2.flag || xMsg1.opt != xMsg2.opt {
		t.Fatal()
	}
	var x *xxx
	err = xMsg2.Unmarshal(&x)
	if err != nil {
		t.Fatal(err)
	}
	if x != nil {
		t.Fatal()
	}
}
