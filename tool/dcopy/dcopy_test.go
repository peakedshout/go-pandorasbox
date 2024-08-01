package dcopy

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"testing"
)

type TestStruct struct {
	A int
	b int

	C   *any
	Sli []string

	M map[string]any

	Ptr *TestStruct
}

func TestCopy(t *testing.T) {
	a := any(1)
	src := &TestStruct{
		A:   100,
		b:   20,
		C:   &a,
		Sli: []string{"1333", "2444"},

		M: map[string]any{"test1121": 1424, "test2141": "xx1xx51515", "13": 33},

		Ptr: &TestStruct{
			A:   200,
			b:   30,
			M:   map[string]any{"tes41411": 5551, "test1142": "xxxx56x", "144": 224},
			Sli: []string{"12331", "55522"},
			Ptr: (*TestStruct)(nil),
		},
	}

	dest := Copy(src)
	if hjson.MustMarshalStr(src) != hjson.MustMarshalStr(dest) {
		t.Fatal("failed")
	}

	u := dest.(*TestStruct)
	u.Ptr.A = 10000
	if u.Ptr.A == src.Ptr.A {
		t.Fatal("failed")
	}
}

func TestCopy1(t *testing.T) {
	str := "123"
	dest := CopyT[string](str)
	if str != dest {
		t.Fatal("failed")
	}
	type xFn func() int
	var fn xFn = func() int {
		return 1
	}
	sFn := fn
	iFn := CopyT[xFn](sFn)
	if sFn() != iFn() {
		t.Fatal("failed")
	}
	fn = func() int {
		return 2
	}
	if fn() == iFn() {
		t.Fatal("failed")
	}
}

type C struct {
	Id int
}
type Info struct {
	A string
	B struct {
		Id int
	}
	C
}

func TestStructAnonymous(t *testing.T) {
	info := Info{
		A: "xx34534535x",
		B: struct{ Id int }{
			Id: 767567567,
		},
		C: C{
			Id: 55555,
		},
	}
	copyInfo := Copy(info)
	if hjson.MustMarshalStr(copyInfo) != hjson.MustMarshalStr(info) {
		t.Fatal("failed")
	}
}

func TestCopySDT(t *testing.T) {
	i1 := 1
	i2 := 2
	CopySDT[int](i1, &i2)
	fmt.Println(i1, i2)
}

func TestCopySDT2(t *testing.T) {
	a := any(1)
	src := &TestStruct{
		A: 100,
		b: 20,
		C: &a,

		M: map[string]any{"test1121": 1424, "test2141": "xx1xx51515", "13": 33},

		Ptr: &TestStruct{
			A:   200,
			b:   30,
			Sli: []string{"12331", "55522"},
			Ptr: (*TestStruct)(nil),
		},
	}
	dest := &TestStruct{
		A:   1,
		b:   20,
		Sli: []string{"1333", "2444", "dwdwdw"},

		Ptr: &TestStruct{
			A:   200,
			b:   30,
			C:   &a,
			M:   map[string]any{"s41411": 551, "tes142": "xxxx56x", "144": 224},
			Sli: []string{"12331", "552"},
		},
	}

	err := CopySDT[*TestStruct](src, &dest)
	if err != nil {
		t.Fatal(err)
	}
	if hjson.MustMarshalStr(src) != hjson.MustMarshalStr(dest) {
		t.Fatal("failed")
	}

	u := dest
	u.Ptr.A = 10000
	if u.Ptr.A == src.Ptr.A {
		t.Fatal("failed")
	}
}
