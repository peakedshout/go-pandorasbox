package htoml

import (
	"github.com/BurntSushi/toml"
)

func Valid(bs []byte) bool {
	m := map[string]any{}
	return toml.Unmarshal(bs, &m) == nil
}

func Unmarshal(in []byte, out interface{}) (err error) {
	return toml.Unmarshal(in, out)
}

func MustUnmarshal(in []byte, out interface{}) {
	err := Unmarshal(in, out)
	if err != nil {
		panic(err)
	}
}

func UnmarshalStr(in string, out interface{}) (err error) {
	return Unmarshal([]byte(in), out)
}

func MustUnmarshalStr(in string, out interface{}) {
	MustUnmarshal([]byte(in), out)
}

func Marshal(in interface{}) (out []byte, err error) {
	return toml.Marshal(in)
}

func MustMarshal(in interface{}) (out []byte) {
	bytes, err := Marshal(in)
	if err != nil {
		panic(err)
	}
	return bytes
}

func MarshalStr(in interface{}) (out string, err error) {
	bytes, err := Marshal(in)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func MustMarshalStr(in interface{}) (out string) {
	return string(MustMarshal(in))
}

func UnmarshalV2(inBody interface{}, outBody interface{}) error {
	b, err := toml.Marshal(inBody)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, outBody)
}

func MustUnmarshalV2(inBody interface{}, outBody interface{}) {
	MustUnmarshal(MustMarshal(inBody), &outBody)
}
