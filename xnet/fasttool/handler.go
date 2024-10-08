package fasttool

import (
	"context"
	"net"
)

type ConnHandler func(ctx context.Context, conn net.Conn) error

type PacketHandler func(pconn net.PacketConn, data []byte, addr net.Addr) error
