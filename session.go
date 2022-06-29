package tcpserve

import (
	"net"
)

// A Codec can encrypt and decrypt packets
type Codec interface {
	Encrypt([]byte) []byte
	Decrypt([]byte) []byte
}

type Session struct {
	id    int
	conn  net.Conn
	codec Codec
}

type SessionOption func(*Session)

func NewSession(id int, conn net.Conn, codec Codec) *Session {
	return &Session{
		id:    id,
		conn:  conn,
		codec: codec,
	}
}

func WithId(id int) SessionOption {
	return func(s *Session) {
		s.id = id
	}
}

func WithConn(conn net.Conn) SessionOption {
	return func(s *Session) {
		s.conn = conn
	}
}

func WithCodec(codec Codec) SessionOption {
	return func(s *Session) {
		s.codec = codec
	}
}

func (s *Session) Id() int {
	return s.id
}

// Encrypt and send a slice of bytes
func (s *Session) Write(data []byte) (int, error) {
	res := s.codec.Encrypt(data)

	return s.conn.Write(res)
}

// Send a slice of bytes (UNENCRYPTED)
func (s *Session) WriteRaw(data []byte) (int, error) {
	return s.conn.Write(data)
}

func (s *Session) Read(data []byte) (int, error) {
	return s.conn.Read(data)
}
