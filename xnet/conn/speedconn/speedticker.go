package speedconn

import (
	"github.com/peakedshout/go-pandorasbox/tool/speed"
	"io"
)

type NetworkSpeedTicker struct {
	rw *speed.ReadWriter
}

func NewNetworkSpeedTicker(rw io.ReadWriter) *NetworkSpeedTicker {
	return &NetworkSpeedTicker{rw: speed.NewReadWriter(rw, rw)}
}

func (nst *NetworkSpeedTicker) transfer(u int, d int) {
	nst.rw.Add(u, d)
}

func (nst *NetworkSpeedTicker) download(p []byte) (n int, err error) {
	return nst.rw.Read(p)
}

func (nst *NetworkSpeedTicker) upload(p []byte) (n int, err error) {
	return nst.rw.Write(p)
}

func (nst *NetworkSpeedTicker) DownloadSpeed() int {
	return nst.rw.RSpeed()
}

func (nst *NetworkSpeedTicker) UploadSpeed() int {
	return nst.rw.WSpeed()
}

func (nst *NetworkSpeedTicker) Speed() (d, u int) {
	return nst.rw.RSpeed(), nst.rw.WSpeed()
}

func (nst *NetworkSpeedTicker) DownloadSpeedView() string {
	return nst.rw.RSpeedView()
}

func (nst *NetworkSpeedTicker) UploadSpeedView() string {
	return nst.rw.WSpeedView()
}

func (nst *NetworkSpeedTicker) SpeedView() (d, u string) {
	return nst.rw.RSpeedView(), nst.rw.WSpeedView()
}
