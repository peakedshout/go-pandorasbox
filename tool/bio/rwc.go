package bio

import "io"

type rwc struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func NewRWC(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return &rwc{
		r: r,
		w: w,
		c: c,
	}
}

func (rwc *rwc) Read(p []byte) (n int, err error) {
	if rwc.r == nil {
		return 0, err
	}
	return rwc.r.Read(p)
}

func (rwc *rwc) Write(p []byte) (n int, err error) {
	if rwc.w == nil {
		return 0, err
	}
	return rwc.w.Write(p)
}

func (rwc *rwc) Close() error {
	if rwc.c == nil {
		return nil
	}
	return rwc.c.Close()
}
