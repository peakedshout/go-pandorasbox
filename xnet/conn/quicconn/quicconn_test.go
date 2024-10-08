package quicconn

import (
	"bytes"
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/peakedshout/go-pandorasbox/tool/mhash"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/quic-go/quic-go"
	"strconv"
	"testing"
	"time"
)

func TestNewConn(t *testing.T) {
	tlsConfig, err := pcrypto.NewDefaultTlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := quic.ListenAddr("127.0.0.1:0", tlsConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
		conn := NewConn(false, connection)
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	connection, err := quic.DialAddr(context.Background(), listener.Addr().String(), tlsConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	conn := NewConn(true, connection)
	defer conn.Close()
	data := []byte(uuid.NewIdn(1024))
	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, buf[:n]) {
			t.Fatal("data fatal")
		}
	}
}

func TestGenKey(t *testing.T) {
	fmt.Println("Crypto Name:", aesgcm.PCryptoAes256Gcm.Name())
	salt := uuid.NewIdn(4096)
	fmt.Println("salt (len 4096):\n", salt)
	key := mhash.ToHash([]byte(salt))
	fmt.Println("key len:", len(key))
	keyStr := make([]string, 0, len(key)*2)
	for _, b := range key {
		keyStr = append(keyStr, strconv.Itoa(int(b)), ",")
	}
	fmt.Println("key:\n", keyStr)
	tu := time.Now().Unix()
	fmt.Println("data (time.Now().Unix()):\n", tu)
	data := []byte(strconv.FormatInt(tu, 10))
	fmt.Println("body len:", len(data))
	encrypt, err := aesgcm.PCryptoAes256Gcm.Encrypt(data, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encrypt len:", len(encrypt))
	decrypt, err := aesgcm.PCryptoAes256Gcm.Decrypt(encrypt, key)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("decrypt len:", len(decrypt))
	parseInt, err := strconv.ParseInt(string(decrypt), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("over (time.Now().Unix()):\n", parseInt)
}
