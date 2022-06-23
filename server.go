package servetcp

import (
	"fmt"
	"net"
)

type IPacket interface {
	Bytes() []byte
	Size() int
}

// An Encrypter is classified as a function that can take in a slice of bytes and return the encryption form of it
type Encrypter func([]byte) []byte

// A Decrypter is classified as a function that can take in a slice of bytes and return the decryption form of it
type Decrypter func([]byte) []byte

// A logger is classified as a function that can take in a string
type Logger func(string)

type TCPServer struct {
	port        int
	log         Logger
	errLog      Logger
	ln          net.Listener
	connections []net.Conn
	countConn   int
}

func NewServer(port int, log Logger) *TCPServer {
	return &TCPServer{
		port: port,
		log:  log,
	}
}

func (s TCPServer) Port() int {
	return s.port
}

func (s *TCPServer) Start() (err error) {
	s.ln, err = net.Listen("tcp", ":8484")
	if err != nil {
		return err
	}
	s.log(fmt.Sprintf("TCP Server started on port %d", s.port))

	// Close listener at end of function
	defer s.ln.Close()

	// Handle each new connection
	for {
		// Block until new connection and accept it
		conn, err := s.ln.Accept()
		if err != nil {
			conn.Close() // Close connection
			s.errLog(fmt.Sprint("error accepting client connection:", err))
			continue // Proceed to block until next client connection
		}

		// Add connection to the slice
		s.connections = append(s.connections, conn)

		defer conn.Close()

		for {
			// Some handling here
		}

	}
}

func (s *TCPServer) Stop() (err error) {
	// Close client connections
	for _, connection := range s.connections {
		connection.Close() // No error handling since we're trying to shut down anyway
	}

	// Close listener
	return s.ln.Close()
}
