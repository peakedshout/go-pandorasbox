package speed

import (
	"io"
)

type ReadWriter struct {
	*Writer
	*Reader
}

func NewReadWriter(r io.Reader, w io.Writer) *ReadWriter {
	return &ReadWriter{
		Writer: NewWriter(w),
		Reader: NewReader(r),
	}
}

func (rw *ReadWriter) Add(w int, r int) {
	rw.Writer.Add(w)
	rw.Reader.Add(r)
}

func (rw *ReadWriter) Speed() (w int, r int) {
	return rw.WSpeed(), rw.RSpeed()
}

func (rw *ReadWriter) SpeedView() (w string, r string) {
	return rw.WSpeedView(), rw.RSpeedView()
}

func (rw *ReadWriter) WSpeed() int {
	return rw.Writer.Speed()
}

func (rw *ReadWriter) WSpeedView() string {
	return rw.Writer.SpeedView()
}

func (rw *ReadWriter) RSpeed() int {
	return rw.Reader.Speed()
}

func (rw *ReadWriter) RSpeedView() string {
	return rw.Reader.SpeedView()
}

type Writer struct {
	w  io.Writer
	st *Ticker
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:  w,
		st: &Ticker{},
	}
}

func (w *Writer) Add(i int) {
	w.st.Set(i)
}

func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err == nil {
		w.Add(n)
	}
	return n, err
}

func (w *Writer) Speed() int {
	return w.st.Get()
}

func (w *Writer) SpeedView() string {
	return formatSpeed(w.Speed())
}

type Reader struct {
	r  io.Reader
	st *Ticker
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:  r,
		st: &Ticker{},
	}
}

func (r *Reader) Add(i int) {
	r.st.Set(i)
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if err == nil {
		r.st.Set(n)
	}
	return n, err
}

func (r *Reader) Speed() int {
	return r.st.Get()
}

func (r *Reader) SpeedView() string {
	return formatSpeed(r.Speed())
}
