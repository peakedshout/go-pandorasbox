package xbit

import (
	"fmt"
	"reflect"
)

type value interface {
	int | uint | uint8 | int8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func Extract[T1, T2 value](a T1, offset, n int) T2 {
	return T2((a >> offset) & (1<<n - 1))
}

func SetFlag[T value](a T, offset int, flag bool) T {
	if flag {
		return T(a | (1 << (offset - 1)))
	} else {
		return T(a &^ (1 << (offset - 1)))
	}
}

func LittleToBytes[T value](t T) []byte {
	tf := reflect.TypeOf(t)
	size := int(tf.Size())
	var bytesFlag = 1<<8 - 1
	bs := make([]byte, size)
	for i := range bs {
		bs[i] = byte((t >> (i * 8)) & T(bytesFlag))
	}
	return bs
}

func BigToBytes[T value](t T) []byte {
	tf := reflect.TypeOf(t)
	size := int(tf.Size())
	var bytesFlag = 1<<8 - 1
	bs := make([]byte, size)
	for i := range bs {
		bs[size-i-1] = byte((t >> (i * 8)) & T(bytesFlag))
	}
	return bs
}

func LittleFromBytes[T value](b []byte) (t T, err error) {
	tf := reflect.TypeOf(t)
	size := int(tf.Size())
	if len(b) != size {
		return t, fmt.Errorf("%T size must be %d not %d", t, size, len(b))
	}
	for i := range b {
		t += T(b[i]) << (i * 8)
	}
	return t, nil
}

func BigFromBytes[T value](b []byte) (t T, err error) {
	tf := reflect.TypeOf(t)
	size := int(tf.Size())
	if len(b) != size {
		return t, fmt.Errorf("%T size must be %d not %d", t, size, len(b))
	}
	for i := range b {
		t += T(b[size-i-1]) << (i * 8)
	}
	return t, nil
}
