package xmsg

import (
	"github.com/peakedshout/go-pandorasbox/protocol"
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"io"
	"sync/atomic"
)

func ReadXMsg(reader io.Reader, cp protocol.Protocol) (xMsg *XMsg, n int, err error) {
	xMsg = new(XMsg)
	var b []byte
	l := 0
	func() {
		if rc, ok := reader.(bio.FlowReader); ok {
			defer rc.Count(&l)()
		}
		err = cp.Decode(reader, &b)
	}()
	if err != nil {
		return nil, 0, err
	}
	if l == 0 {
		l = len(b)
	}
	err = xMsg.unmarshal(b)
	if err != nil {
		return nil, 0, err
	}
	return xMsg, l, nil
}

func WriteCMsg(writer io.Writer, cp protocol.Protocol, xMsg *XMsg) (int, error) {
	b, err := xMsg.marshal()
	if err != nil {
		return 0, err
	}
	l := len(b)
	if wc, ok := writer.(bio.FlowWriter); ok {
		l = 0
		defer wc.Count(&l)()
	}
	err = cp.Encode(writer, b)
	return l, err
}

type XLauncher interface {
	XReadLauncher
	XWriteLauncher
}

type XReadLauncher interface {
	ReadXMsg() (xMsg *XMsg, n int, err error)
}
type XWriteLauncher interface {
	SendXMsg(header string, id uint32, opt OptType, data any) (uint32, int, error)
	RecvXMsg(header string, id uint32, opt OptType, data any) (uint32, int, error)
}

func NewXReadLauncher(reader io.Reader, cp protocol.Protocol) XReadLauncher {
	return &xReadLauncher{
		reader: reader,
		cp:     cp,
	}
}

type xReadLauncher struct {
	reader io.Reader
	cp     protocol.Protocol
}

func (r *xReadLauncher) ReadXMsg() (xMsg *XMsg, n int, err error) {
	xMsg = new(XMsg)
	var b []byte
	l := 0
	func() {
		if rc, ok := r.reader.(bio.FlowReader); ok {
			defer rc.Count(&l)()
		}
		err = r.cp.Decode(r.reader, &b)
	}()
	if err != nil {
		return nil, 0, err
	}
	if l == 0 {
		l = len(b)
	}
	err = xMsg.unmarshal(b)
	if err != nil {
		return nil, 0, err
	}
	return xMsg, l, nil
}

func NewXWriteLauncher(writer io.Writer, cp protocol.Protocol, flag flagEnum) XWriteLauncher {
	return &xWriteLauncher{
		writer: writer,
		cp:     cp,
		flag:   flag,
	}
}

type xWriteLauncher struct {
	writer io.Writer
	cp     protocol.Protocol
	flag   flagEnum
	id     uint32
}

func (x *xWriteLauncher) SendXMsg(header string, id uint32, opt OptType, data any) (xid uint32, n int, err error) {
	if id == 0 {
		id = x.getId()
	}
	xMsg, err := newXMsg(header, x.flag, id, opt, data)
	if err != nil {
		return 0, 0, err
	}
	b, err := xMsg.marshal()
	if err != nil {
		return 0, 0, err
	}
	l := len(b)
	if wc, ok := x.writer.(bio.FlowWriter); ok {
		l = 0
		defer wc.Count(&l)()
	}
	err = x.cp.Encode(x.writer, b)
	if err != nil {
		return 0, 0, err
	}
	return id, l, nil
}

func (x *xWriteLauncher) RecvXMsg(header string, id uint32, opt OptType, data any) (xid uint32, n int, err error) {
	if id == 0 {
		id = x.getId()
	}
	xMsg, err := newXMsg(header, x.flag^1, id, opt, data)
	if err != nil {
		return 0, 0, err
	}
	b, err := xMsg.marshal()
	if err != nil {
		return 0, 0, err
	}
	l := len(b)
	if wc, ok := x.writer.(bio.FlowWriter); ok {
		l = 0
		defer wc.Count(&l)()
	}
	err = x.cp.Encode(x.writer, b)
	if err != nil {
		return 0, 0, err
	}
	return id, l, nil
}

func (x *xWriteLauncher) getId() uint32 {
	for id := atomic.AddUint32(&x.id, 1); ; {
		if id != 0 {
			return id
		}
	}
}

func NewXLauncher(rw io.ReadWriter, cp protocol.Protocol, flag flagEnum) XLauncher {
	return &xLauncher{
		XReadLauncher:  NewXReadLauncher(rw, cp),
		XWriteLauncher: NewXWriteLauncher(rw, cp, flag),
	}
}

type xLauncher struct {
	XReadLauncher
	XWriteLauncher
}
