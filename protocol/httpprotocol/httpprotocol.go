package httpprotocol

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/peakedshout/go-pandorasbox/protocol"
	"github.com/peakedshout/go-pandorasbox/protocol/jsonprotocol"
	"io"
	"net/http"
	"net/url"
)

func NewHttpProtocol(client bool, p protocol.Protocol, cECb, sDCb func(*http.Request) error, cDCb, sECb func(*http.Response) error) protocol.Protocol {
	hp := &httpProtocol{
		isClient: client,
		cECb:     cECb,
		sECb:     sECb,
		sDCb:     sDCb,
		cDCb:     cDCb,
		p:        p,
	}
	if hp.p == nil {
		hp.p = new(jsonprotocol.JsonProtocol)
	}
	return hp
}

type httpProtocol struct {
	isClient bool
	cECb     func(req *http.Request) error
	sECb     func(resp *http.Response) error
	sDCb     func(req *http.Request) error
	cDCb     func(resp *http.Response) error
	p        protocol.Protocol
}

func (h *httpProtocol) Encode(w io.Writer, a any) error {
	bs := new(bytes.Buffer)
	err := h.p.Encode(bs, a)
	if err != nil {
		return err
	}
	if !h.isClient {
		response := &http.Response{
			StatusCode:    http.StatusOK,
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: 0,
		}
		if h.sECb != nil {
			err = h.sECb(response)
			if err != nil {
				return err
			}
		}
		response.Body = io.NopCloser(bs)
		response.ContentLength = int64(bs.Len())
		return response.Write(w)
	} else {
		u, _ := url.Parse("http://0.0.0.0")
		request := &http.Request{
			Method:        http.MethodGet,
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: 0,
			URL:           u,
			Host:          "0.0.0.0",
		}
		if h.cECb != nil {
			err = h.cECb(request)
			if err != nil {
				return err
			}
		}
		request.Body = io.NopCloser(bs)
		request.ContentLength = int64(bs.Len())
		return request.Write(w)
	}
}

func (h *httpProtocol) Decode(r io.Reader, a any) error {
	reader := bufio.NewReader(r)
	if !h.isClient {
		request, err := http.ReadRequest(reader)
		if err != nil {
			return err
		}
		if request.Body == nil {
			return errors.New("nil body")
		}
		if h.sDCb != nil {
			err = h.sDCb(request)
			if err != nil {
				return err
			}
		}
		err = h.p.Decode(request.Body, a)
		_ = request.Body.Close()
		if err != nil {
			return err
		}
	} else {
		response, err := http.ReadResponse(reader, nil)
		if err != nil {
			return err
		}
		if response.Body == nil {
			return errors.New("nil body")
		}
		if h.cDCb != nil {
			err = h.cDCb(response)
			if err != nil {
				return err
			}
		}
		err = h.p.Decode(response.Body, a)
		_ = response.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
