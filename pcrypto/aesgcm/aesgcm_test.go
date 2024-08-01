package aesgcm

import (
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func Test(t *testing.T) {
	key := []byte(uuid.NewId(1))
	b, err := PCryptoAes256Gcm.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAes256Gcm.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}

	// hash
	key = []byte(uuid.NewId(1)[5:])
	b, err = PCryptoAesHashGcm.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAesHashGcm.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}
}
