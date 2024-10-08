package cfcprotocol

import (
	"encoding/json"
	"io"
	"reflect"
)

type msgType uint8

const (
	msgTypeRawBytes = msgType(iota)
	msgTypeJsonBytes
)

func encode(a any) ([]byte, error) {
	switch v := a.(type) {
	case []byte:
		return append([]byte{byte(msgTypeRawBytes)}, v...), nil
	case io.Reader:
		all, err := io.ReadAll(a.(io.Reader))
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(msgTypeRawBytes)}, all...), nil
	default:
		jb, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(msgTypeJsonBytes)}, jb...), nil
	}
}

func decode(b []byte, a any) error {
	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return ErrCFCProtocolDecodeToNonNilPointer.Errorf(rv.Type().Name())
	}
	if len(b) == 0 {
		return ErrCFCProtocolDecodeNilData.Errorf()
	}
	switch msgType(b[0]) {
	case msgTypeRawBytes:
		switch a.(type) {
		case *[]byte:
			bs := a.(*[]byte)
			*bs = b[1:]
			return nil
		case io.Writer:
			w := a.(io.Writer)
			_, err := w.Write(b[1:])
			if err != nil {
				return err
			}
			return nil
		default:
			return ErrCFCProtocolDecodeInvalidContainer.Errorf()
		}
	case msgTypeJsonBytes:
		return json.Unmarshal(b[1:], a)
	default:
		return ErrCFCProtocolDecodeInvalidMsgType.Errorf()
	}
}
