package uerror

import (
	"testing"
)

func TestMuxErrorCode(t *testing.T) {
	_muxDebug = true
}

func TestUnmarshalUError(t *testing.T) {
	ec1 := NewErrorCode(0, 0, "%v")
	ec2 := NewErrorCode(0, 1, "%v")
	err2 := ec2.Errorf("123")
	err1 := ec1.Errorf(err2)

	uError, b := UnmarshalUError(err1.Error())
	if !b || uError.wrapError == nil {
		t.Fatal()
	}
	u, ok := uError.wrapError.(*UError)
	if !ok || u.Code() != ec2.Code() {
		t.Fatal()
	}
}
