package xerror

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	te := New("xxxx")
	tec := New("xxxx")
	if !errors.Is(te, tec) {
		t.Fatal()
	}
	e1 := New("ccc %w")
	e2 := e1.Errorf(te)
	if !errors.Is(e1, e2) {
		t.Fatal()
	}
	if !errors.Is(e2, te) {
		t.Fatal()
	}
	fmt.Println(e1, e2)
	if !errors.Is(errors.Unwrap(e2), tec) {
		t.Fatal()
	}
	xe := new(*XError)
	if !errors.As(e2, xe) {
		t.Fatal()
	}
	fmt.Println(*xe)
}
