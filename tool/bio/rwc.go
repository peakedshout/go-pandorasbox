package bio

import "io"

type _rwc struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func NewRWC(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	return &_rwc{
		r: r,
		w: w,
		c: c,
	}
}

func (rwc *_rwc) Read(p []byte) (n int, err error) {
	if rwc.r == nil {
		return 0, err
	}
	return rwc.r.Read(p)
}

func (rwc *_rwc) Write(p []byte) (n int, err error) {
	if rwc.w == nil {
		return 0, err
	}
	return rwc.w.Write(p)
}

func (rwc *_rwc) Close() error {
	if rwc.c == nil {
		return nil
	}
	return rwc.c.Close()
}
