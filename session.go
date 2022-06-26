package tcpserve

import (
	"net"
)

// A Codec is classified as a function that can take in a slice of bytes and return the manipulated form of it
type Codec func([]byte)

type Session struct {
	id      int
	conn    net.Conn
	encrypt Codec
	decrypt Codec
}

func NewSession(id int, conn net.Conn) *Session {
	return &Session{
		id:      id,
		conn:    conn,
		encrypt: func([]byte) {}, decrypt: func([]byte) {},
	}
}

func (s *Session) Id() int {
	return s.id
}

// SetEncrypter changes the session's encrypter
func (s *Session) SetEncrypter(encrypter Codec) {
	s.encrypt = encrypter
}

// SetDecrypter changes the session's decrypter
func (s *Session) SetDecrypter(decrypter Codec) {
	s.decrypt = decrypter
}

// Encrypt and send a slice of bytes
func (s *Session) Write(data []byte) (int, error) {
	s.encrypt(data)

	return s.conn.Write(data)
}

// Send a slice of bytes (UNENCRYPTED)
func (s *Session) WriteRaw(data []byte) (int, error) {
	return s.conn.Write(data)
}

func (s *Session) Read(data []byte) (int, error) {
	return s.conn.Read(data)
}