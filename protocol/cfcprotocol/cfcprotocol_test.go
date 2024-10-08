package cfcprotocol

import (
	"bufio"
	"bytes"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/pcrypto/rsa"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"strings"
	"testing"
)

func TestBytes(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	data := uuid.NewId(1024)
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(&bf)
	var s string

	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
}

func TestBytes2(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	data := []byte(uuid.NewId(1024))
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(&bf)
	var s []byte

	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s, data) {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s, data) {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s, data) {
		t.Failed()
	}
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s, data) {
		t.Failed()
	}
}

func TestBytes3(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	rs := uuid.NewIdn(1024)
	data := strings.NewReader(rs)
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	data.Reset(rs)
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	data.Reset(rs)
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	data.Reset(rs)
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	reader := bufio.NewReader(&bf)
	var s bytes.Buffer

	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	eb := []byte(rs)
	if bytes.Equal(s.Bytes(), eb) {
		t.Failed()
	}
	s.Reset()
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s.Bytes(), eb) {
		t.Failed()
	}
	s.Reset()
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s.Bytes(), eb) {
		t.Failed()
	}
	s.Reset()
	err = cp.Decode(reader, &s)
	if err != nil {
		t.Error(err)
	}
	if bytes.Equal(s.Bytes(), eb) {
		t.Failed()
	}
}

func TestNewCFCProtocol(t *testing.T) {
	key := []byte("00000000000000000000000000000000")
	pc, err := pcrypto.NewPCrypto(aesgcm.PCryptoAes256Gcm, key)
	if err != nil {
		t.Error(err)
	}
	data := "hello,world!"
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	var s string
	err = cp.Decode(&bf, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
}

func TestNewCFCProtocol2(t *testing.T) {
	pub, pri, err := rsa.PCryptoRsa.GenRsaKey(1024)
	if err != nil {
		t.Error(err)
	}
	pc, err := pcrypto.NewPCrypto(rsa.PCryptoRsa, pub, pri)
	if err != nil {
		t.Error(err)
	}

	data := "hello,world!"
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	var s string
	err = cp.Decode(&bf, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
}

func TestNewCFCProtocol3(t *testing.T) {
	pc, err := pcrypto.NewPCrypto(nil)
	if err != nil {
		t.Error(err)
	}

	data := "hello,world!"
	cp := NewCFCProtocol(pc)
	var bf bytes.Buffer
	err = cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	var s string
	err = cp.Decode(&bf, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
}

func TestNewCFCProtocol4(t *testing.T) {
	data := "hello,world!"
	cp := NewCFCProtocol(pcrypto.CryptoPlaintext)
	var bf bytes.Buffer
	err := cp.Encode(&bf, data)
	if err != nil {
		t.Error(err)
	}
	var s string
	err = cp.Decode(&bf, &s)
	if err != nil {
		t.Error(err)
	}
	if s != data {
		t.Failed()
	}
}
