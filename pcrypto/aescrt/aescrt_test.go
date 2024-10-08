package aescrt

import (
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func Test(t *testing.T) {
	key := []byte(uuid.NewId(1))
	b, err := PCryptoAes256Crt.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAes256Crt.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}

	// hash
	key = []byte(uuid.NewId(1)[5:])
	b, err = PCryptoAesHashCrt.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAesHashCrt.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}
}
