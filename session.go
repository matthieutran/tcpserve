package tcpserve

import (
	"net"
)

// A Codec performs operations on an input byte slice and returns the result
type Codec func([]byte) []byte

type Session struct {
	id      int
	conn    net.Conn
	encrypt Codec
	decrypt Codec
}

type SessionOption func(*Session)

func NewSession(options ...SessionOption) *Session {
	s := &Session{}
	dummy := func(b []byte) []byte {
		return b
	}

	s.encrypt = dummy
	s.decrypt = dummy

	// Call each option
	for _, option := range options {
		option(s)
	}

	return s
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

func WithEncrypter(encrypter Codec) SessionOption {
	return func(s *Session) {
		s.encrypt = encrypter
	}
}

func WithDecrypter(decrypter Codec) SessionOption {
	return func(s *Session) {
		s.decrypt = decrypter
	}
}

func (s *Session) Id() int {
	return s.id
}

func (s *Session) Encrypt(data []byte) []byte {
	return s.encrypt(data)
}

func (s *Session) Decrypt(data []byte) []byte {
	return s.decrypt(data)
}

// Encrypt and send a slice of bytes
func (s *Session) Write(data []byte) (int, error) {
	res := s.Encrypt(data)

	return s.conn.Write(res)
}

// Send a slice of bytes (UNENCRYPTED)
func (s *Session) WriteRaw(data []byte) (int, error) {
	return s.conn.Write(data)
}

func (s *Session) Read(data []byte) (int, error) {
	return s.conn.Read(data)
}
