package bio

import (
	"bytes"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func TestFlow(t *testing.T) {
	bs := new(bytes.Buffer)
	w := NewFlowWriter(bs)
	wi := 0
	wc := w.Count(&wi)
	data := []byte(uuid.NewIdn(4096))
	for i := 0; i < 100; i++ {
		_, _ = w.Write(data)
	}
	wc()
	if wi != 4096*100 {
		t.Fail()
	}
	r := NewFlowReader(bs)
	ri := 0
	rc := r.Count(&ri)
	for i := 0; i < 100; i++ {
		_, _ = r.Read(data)
	}
	rc()
	if ri != wi {
		t.Fail()
	}
}
