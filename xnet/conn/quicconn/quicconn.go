package quicconn

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto/aesgcm"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

var _quicConnHandShakeHandMaxDuration = 60 * time.Second

type QuicConn struct {
	isClient bool
	conn     quic.Connection
	stream   quic.Stream
	once     sync.Once
	mux      sync.Mutex
	timeout  time.Duration
	herr     error
}

func (q *QuicConn) Read(b []byte) (n int, err error) {
	err = q.HandShakeHandle()
	if err != nil {
		return 0, err
	}
	return q.stream.Read(b)
}

func (q *QuicConn) Write(b []byte) (n int, err error) {
	err = q.HandShakeHandle()
	if err != nil {
		return 0, err
	}
	return q.stream.Write(b)
}

func (q *QuicConn) Close() (err error) {
	q.mux.Lock()
	defer q.mux.Unlock()
	return q.close()
}

func (q *QuicConn) LocalAddr() net.Addr {
	return q.conn.LocalAddr()
}

func (q *QuicConn) RemoteAddr() net.Addr {
	return q.conn.RemoteAddr()
}

func (q *QuicConn) SetDeadline(t time.Time) (err error) {
	err = q.HandShakeHandle()
	if err != nil {
		return err
	}
	return q.stream.SetDeadline(t)
}

func (q *QuicConn) SetReadDeadline(t time.Time) (err error) {
	err = q.HandShakeHandle()
	if err != nil {
		return err
	}
	return q.stream.SetReadDeadline(t)
}

func (q *QuicConn) SetWriteDeadline(t time.Time) (err error) {
	err = q.HandShakeHandle()
	if err != nil {
		return err
	}
	return q.stream.SetWriteDeadline(t)
}

// HandShakeHandle activate stream and check peer time duration
func (q *QuicConn) HandShakeHandle() error {
	q.once.Do(func() {
		q.mux.Lock()
		defer q.mux.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), q.timeout)
		defer cancel()
		if q.isClient {
			q.herr = q.handShakeClient(ctx)
			if q.herr != nil {
				q.close()
				q.herr = fmt.Errorf("client handshake handle bad:%w", q.herr)
			}
		} else {
			q.herr = q.handShakeServer(ctx)
			if q.herr != nil {
				q.close()
				q.herr = fmt.Errorf("server handshake handle bad:%w", q.herr)
			}
		}
	})
	return q.herr
}

func (q *QuicConn) close() (err error) {
	if q.stream != nil {
		err = q.stream.Close()
	}
	err1 := q.conn.CloseWithError(400, "closed")
	if err != nil {
		return err
	}
	return err1
}

func (q *QuicConn) handShakeClient(ctx context.Context) (err error) {
	stream, err := q.conn.AcceptStream(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()
	t := time.Now()
	buf := make([]byte, _decryptlen)
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return err
	}
	err = checkHandShakeData(t, buf)
	if err != nil {
		return err
	}
	_, err = stream.Write(makeHandShakeData(t))
	if err != nil {
		return err
	}
	q.stream = stream
	return nil
}

func (q *QuicConn) handShakeServer(ctx context.Context) (err error) {
	stream, err := q.conn.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()
	t := time.Now()
	_, err = stream.Write(makeHandShakeData(t))
	if err != nil {
		return err
	}
	buf := make([]byte, _decryptlen)
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return err
	}
	err = checkHandShakeData(t, buf)
	if err != nil {
		return err
	}
	q.stream = stream
	return nil
}

func checkHandShakeData(t time.Time, data []byte) error {
	decrypt, err := aesgcm.PCryptoAes256Gcm.Decrypt(data, _genKey)
	if err != nil {
		return err
	}
	rt, err := getTimeUnix(decrypt)
	if err != nil {
		return err
	}
	td := t.Sub(rt)
	if td > _quicConnHandShakeHandMaxDuration || td < -_quicConnHandShakeHandMaxDuration {
		return errors.New("time unix duration invalid")
	}
	return nil
}

func makeHandShakeData(t time.Time) []byte {
	unix := makeTimeUnix(t)
	if len(unix) != _bodylen {
		panic(fmt.Sprintf("body len is not %d", _bodylen))
	}
	encrypt, err := aesgcm.PCryptoAes256Gcm.Encrypt(unix, _genKey)
	if err != nil {
		panic(err)
	}
	if len(encrypt) != _encryptlen {
		panic(fmt.Sprintf("encrypt len is not %d", _encryptlen))
	}
	return encrypt
}

func getTimeUnix(data []byte) (time.Time, error) {
	parseInt, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(parseInt, 0), err
}

func makeTimeUnix(t time.Time) []byte {
	return []byte(strconv.FormatInt(t.Unix(), 10))
}

func NewConn(isClient bool, conn quic.Connection) net.Conn {
	wc := &QuicConn{
		isClient: isClient,
		conn:     conn,
		timeout:  30 * time.Second,
	}
	return wc
}
