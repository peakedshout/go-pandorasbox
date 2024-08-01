package bio

import (
	"io"
	"sync"
)

func NewFlowReadWriter(rw io.ReadWriter) FlowReadWriter {
	return &flowReadWriter{
		fr: NewFlowReader(rw),
		fw: NewFlowWriter(rw),
	}
}

func NewFlowWriter(w io.Writer) FlowWriter {
	return &flowWriter{
		Writer: w,
		mux:    sync.Mutex{},
		count:  nil,
	}
}

func NewFlowReader(r io.Reader) FlowReader {
	return &flowReader{
		Reader: r,
		mux:    sync.Mutex{},
		count:  nil,
	}
}

type FlowReadWriter interface {
	io.ReadWriter
	ReadCount(*int) func()
	WriteCount(*int) func()
}

type FlowWriter interface {
	io.Writer
	Count(*int) func()
}

type FlowReader interface {
	io.Reader
	Count(*int) func()
}

type flowReadWriter struct {
	fr FlowReader
	fw FlowWriter
}

func (f *flowReadWriter) Read(p []byte) (n int, err error) {
	return f.fr.Read(p)
}

func (f *flowReadWriter) Write(p []byte) (n int, err error) {
	return f.fw.Write(p)
}

func (f *flowReadWriter) ReadCount(i *int) func() {
	return f.fr.Count(i)
}

func (f *flowReadWriter) WriteCount(i *int) func() {
	return f.fw.Count(i)
}

type flowWriter struct {
	io.Writer
	mux   sync.Mutex
	count *int
}

func (fw *flowWriter) Write(p []byte) (n int, err error) {
	n, err = fw.Writer.Write(p)
	if n != 0 {
		fw.mux.Lock()
		defer fw.mux.Unlock()
		if fw.count != nil {
			*fw.count += n
		}
	}
	return n, err
}

func (fw *flowWriter) Count(i *int) func() {
	fw.mux.Lock()
	defer fw.mux.Unlock()
	fw.count = i
	return func() {
		fw.mux.Lock()
		defer fw.mux.Unlock()
		fw.count = nil
	}
}

type flowReader struct {
	io.Reader
	mux   sync.Mutex
	count *int
}

func (fr *flowReader) Read(p []byte) (n int, err error) {
	n, err = fr.Reader.Read(p)
	if n != 0 {
		fr.mux.Lock()
		defer fr.mux.Unlock()
		if fr.count != nil {
			*fr.count += n
		}
	}
	return n, err
}

func (fr *flowReader) Count(i *int) func() {
	fr.mux.Lock()
	defer fr.mux.Unlock()
	fr.count = i
	return func() {
		fr.mux.Lock()
		defer fr.mux.Unlock()
		fr.count = nil
	}
}
