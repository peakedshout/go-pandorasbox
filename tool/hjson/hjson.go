package hjson

import (
	"encoding/json"
)

func MarshalStr(body interface{}) (string, error) {
	bytes, err := json.Marshal(body)
	if err != nil {
		return "", nil
	}
	return string(bytes), nil
}

func UnmarshalStr(str string, body interface{}) error {
	return json.Unmarshal([]byte(str), body)
}

func MustUnmarshal(b []byte, body interface{}) {
	if len(b) == 0 {
		return
	}
	err := json.Unmarshal(b, &body)
	if err != nil {
		panic(err)
	}
}

func MustUnmarshalStr(str string, body interface{}) {
	if len(str) == 0 {
		return
	}
	err := json.Unmarshal([]byte(str), &body)
	if err != nil {
		panic(err)
	}
}

func MustMarshal(body interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return b
}

func MustMarshalStr(body interface{}) string {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func UnmarshalV2(inBody interface{}, outBody interface{}) error {
	b, err := json.Marshal(inBody)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, outBody)
}

func MustUnmarshalV2(inBody interface{}, outBody interface{}) {
	MustUnmarshal(MustMarshal(inBody), &outBody)
}

func MarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "\t")
}

func MustMarshalIndent(v any) []byte {
	b, err := MarshalIndent(v)
	if err != nil {
		panic(err)
	}
	return b
}
