package aesofb

import (
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"testing"
)

func Test(t *testing.T) {
	key := []byte(uuid.NewId(1))
	b, err := PCryptoAes256Ofb.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAes256Ofb.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}

	// hash
	key = []byte(uuid.NewId(1)[5:])
	b, err = PCryptoAesHashOfb.Encrypt([]byte("hello!"), key)
	if err != nil {
		t.Error(err)
	}
	b, err = PCryptoAesHashOfb.Decrypt(b, key)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "hello!" {
		t.Failed()
	}
}
