package dnsproxy

import (
	"context"
	"errors"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
)

func NewClient(reqCb ReqCb) (*Client, error) {
	if reqCb == nil {
		return nil, errors.New("nil resp call back")
	}
	return &Client{
		reqCb: reqCb,
	}, nil
}

type Client struct {
	reqCb ReqCb
	mux   sync.Mutex
	uid   uint16
}

type ReqCb func(ctx context.Context, message *dnsmessage.Message) (*dnsmessage.Message, error)

func (c *Client) RawSend(message *dnsmessage.Message) (*dnsmessage.Message, error) {
	return c.RawSendContext(context.Background(), message)
}

func (c *Client) RawSendContext(ctx context.Context, message *dnsmessage.Message) (*dnsmessage.Message, error) {
	return c.reqCb(ctx, message)
}

func (c *Client) LookupIP(name string) ([]string, error) {
	return c.LookupIPContext(context.Background(), name)
}

func (c *Client) LookupIPContext(ctx context.Context, name string) ([]string, error) {
	dname, err := dnsmessage.NewName(name + ".")
	if err != nil {
		return nil, err
	}
	id := c.newId()
	message := dnsmessage.Message{
		Header: dnsmessage.Header{
			ID: id,
		},
		Questions: []dnsmessage.Question{{
			Name:  dname,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
		}},
	}
	resp, err := c.reqCb(ctx, &message)
	if err != nil {
		return nil, err
	}
	if !resp.Response {
		return nil, errors.New("nil response")
	}
	if resp.RCode != dnsmessage.RCodeSuccess {
		return nil, errors.New(resp.RCode.String())
	}
	sl := make([]string, 0, len(resp.Answers))
	for _, answer := range resp.Answers {
		switch answer.Body.(type) {
		case *dnsmessage.AResource:
			r := answer.Body.(*dnsmessage.AResource)
			sl = append(sl, net.IP(r.A[:]).String())
		case *dnsmessage.AAAAResource:
			r := answer.Body.(*dnsmessage.AAAAResource)
			sl = append(sl, net.IP(r.AAAA[:]).String())
		case *dnsmessage.TXTResource:
			r := answer.Body.(*dnsmessage.TXTResource)
			sl = append(sl, r.TXT...)
		default:
			return nil, errors.New("nil response")
		}
	}
	return sl, nil
}

func (c *Client) newId() uint16 {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.uid++
	return c.uid
}
