package xmsg

import (
	"bytes"
	"encoding/json"
	"github.com/peakedshout/go-pandorasbox/tool/xbit"
	"io"
	"reflect"
)

// NoBytes Use it to wrap raw json data instead of being wrapped into bytes for transmission.
type NoBytes []byte

type OptType byte

type flagEnum byte

const (
	FlagZero flagEnum = iota
	FlagOne
)

type dataType byte

const (
	dataTypeRawBytes dataType = iota
	dataTypeStringBytes
	dataTypeJsonBytes
	dataTypeErrorBytes
)

func newXMsg(header string, flag flagEnum, id uint32, opt OptType, data any) (*XMsg, error) {
	if len(header) > 255 {
		return nil, ErrHeaderMustBeLessEqual
	}
	xMsg := &XMsg{
		header: header,
		flag:   flag,
		id:     id,
		opt:    opt,
	}
	if data != nil {
		err := xMsg.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	return xMsg, nil
}

type XMsg struct {
	header string   // len must be <= 255
	flag   flagEnum // 1 or 0 to id
	id     uint32   // 0 ~ (1 << 31 -1)
	opt    OptType
	data   []byte // json bytes , string bytes or raw bytes , set ptr if only use json byte
}

func (x *XMsg) Header() string {
	if x == nil {
		return ""
	}
	return x.header
}

func (x *XMsg) Opt() OptType {
	if x == nil {
		return OptUnknown
	}
	return x.opt
}

func (x *XMsg) Id() uint32 {
	if x == nil {
		return 0
	}
	return x.id
}

func (x *XMsg) Flag() flagEnum {
	if x == nil {
		return 0
	}
	return x.flag
}

func (x *XMsg) NilData() bool {
	return len(x.data) == 0
}

// Unmarshal If the transmitted content is of type error, then it will not be filled out, but will be returned from the function.
func (x *XMsg) Unmarshal(out any) (err error) {
	rv := reflect.ValueOf(out)
	if len(x.data) == 0 {
		return ErrDataOutputNotData
	}
	switch dataType(x.data[0]) {
	case dataTypeRawBytes:
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return ErrDataOutputToNonNilPointer.Errorf(rv.Type().Name())
		}
		switch v := out.(type) {
		case *[]byte:
			o := v
			*o = x.data[1:]
			return nil
		case io.Writer:
			w := out.(io.Writer)
			_, err = w.Write(x.data[1:])
			return err
		default:
			return ErrDataOutputTypeInvalid
		}
	case dataTypeStringBytes:
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return ErrDataOutputToNonNilPointer.Errorf(rv.Type().Name())
		}
		if o, ok := out.(*string); ok {
			*o = string(x.data[1:])
			return nil
		} else {
			return ErrDataOutputTypeInvalid
		}
	case dataTypeJsonBytes:
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return ErrDataOutputToNonNilPointer.Errorf(rv.Type().Name())
		}
		if o, ok := out.(*NoBytes); ok {
			*o = x.data[1:]
			return nil
		}
		err = json.Unmarshal(x.data[1:], out)
		if err != nil {
			return err
		}
		return nil
	case dataTypeErrorBytes:
		return ErrDataOutputError.Errorf(string(x.data[1:]))
	default:
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return ErrDataOutputToNonNilPointer.Errorf(rv.Type().Name())
		}
		return ErrDataOutputTypeInvalid
	}
}

// Marshal If the content of the transmission is of type error, then the error will be transmitted in a special form, and the deserialization side will get an error. If you don't want to do that, you should pass err.Error().
func (x *XMsg) Marshal(data any) error {
	if data == nil {
		x.data = nil
		return nil
	}
	bs := new(bytes.Buffer)
	switch v := data.(type) {
	case []byte:
		bs.WriteByte(byte(dataTypeRawBytes))
		bs.Write(v)
	case io.Reader:
		all, err := io.ReadAll(data.(io.Reader))
		if err != nil {
			return err
		}
		bs.WriteByte(byte(dataTypeRawBytes))
		bs.Write(all)
	case string:
		bs.WriteByte(byte(dataTypeStringBytes))
		bs.Write([]byte(data.(string)))
	case error:
		bs.WriteByte(byte(dataTypeErrorBytes))
		bs.Write([]byte(data.(error).Error()))
	case NoBytes:
		bs.WriteByte(byte(dataTypeJsonBytes))
		bs.Write(data.(NoBytes))
	default:
		jb, err := json.Marshal(data)
		if err != nil {
			return err
		}
		bs.WriteByte(byte(dataTypeJsonBytes))
		bs.Write(jb)
	}
	x.data = bs.Bytes()
	return nil
}

func (x *XMsg) marshal() ([]byte, error) {
	hl := len(x.header)
	if hl > 255 {
		return nil, ErrHeaderMustBeLessEqual
	}
	bs := new(bytes.Buffer)
	bs.WriteByte(byte(hl))
	bs.WriteString(x.header)
	bs.Write(xbit.BigToBytes[uint32](x.mask()))
	bs.Write(xbit.BigToBytes[byte](byte(x.opt)))
	if len(x.data) != 0 {
		bs.Write(x.data)
		return bs.Bytes(), nil
	}
	return bs.Bytes(), nil
}

func (x *XMsg) unmarshal(b []byte) error {
	reader := bytes.NewReader(b)
	readByte, err := reader.ReadByte()
	if err != nil {
		return err
	}
	buf := make([]byte, int(readByte))
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	x.header = string(buf)
	buf = make([]byte, 4)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	id, err := xbit.BigFromBytes[uint32](buf)
	if err != nil {
		return err
	}
	x.unmask(id)
	opt, err := reader.ReadByte()
	if err != nil {
		return err
	}
	x.opt = OptType(opt)
	all, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	x.data = all
	return nil
}

const lower31BitMask uint32 = 1 << 31

func (x *XMsg) mask() uint32 {
	return (x.id & ^lower31BitMask) | (uint32(x.flag) << 31)
}

func (x *XMsg) unmask(id uint32) {
	x.flag = flagEnum(id >> 31)
	x.id = id &^ (lower31BitMask)
}
