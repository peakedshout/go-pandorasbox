package hyaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"testing"
)

type XXX struct {
	A string       `yaml:"a" comment:"l-A" hcomment:"h-A" fcomment:"f-A"`
	B int          `yaml:"b" comment:"l-B" hcomment:"h-B" fcomment:"f-B"`
	C bool         `yaml:"c" comment:"l-C" hcomment:"h-C" fcomment:"f-C"`
	D []XXX1       `yaml:"d" comment:"l-D" hcomment:"h-D" fcomment:"f-D"`
	E *XXX1        `yaml:"e" comment:"l-E" hcomment:"h-E" fcomment:"f-E"`
	F map[int]XXX1 `yaml:"f" comment:"l-F" hcomment:"h-F" fcomment:"f-F"`
}

type XXX1 struct {
	A1 string            `yaml:"a1" comment:"l1-A" hcomment:"h1-A" fcomment:"f1-A"`
	B2 map[string]string `yaml:"b2" comment:"l1-B" hcomment:"h1-B" fcomment:"f1-B"`
}

func TestMustMarshalWithCommentStr(t *testing.T) {
	x := XXX{
		A: "123",
		B: 456,
		C: true,
		D: []XXX1{{
			A1: "dddddd",
			B2: map[string]string{
				"ddff":    "ffff",
				"222ddff": "222ffff",
			},
		}, {
			A1: "dddddd",
			B2: map[string]string{
				"ddff":    "ffff",
				"222ddff": "222ffff",
			},
		}},
		E: &XXX1{
			A1: "xxx333",
			B2: map[string]string{
				"ddff":    "ffff",
				"222ddff": "222ffff",
			},
		},
		F: map[int]XXX1{
			221: {
				A1: "ccc",
				B2: map[string]string{
					"ddff":    "ffff",
					"222ddff": "222ffff",
				},
			},
		},
	}
	out, err := yaml.Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(out))

	fmt.Println(MustMarshalWithCommentStr(x))
}
