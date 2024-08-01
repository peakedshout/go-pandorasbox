package aescbc

import (
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func Test(t *testing.T) {
	key := []byte(uuid.NewId(1))
	b, err := PCryptoAes256Cbc.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAes256Cbc.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}

	// hash
	key = []byte(uuid.NewId(1)[5:])
	b, err = PCryptoAesHashCbc.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAesHashCbc.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}
}
