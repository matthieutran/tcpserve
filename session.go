package tcpserve

import "net"

type Session struct {
	conn net.Conn
}
