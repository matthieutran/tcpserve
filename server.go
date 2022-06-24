package tcpserve

import (
	"fmt"
	"net"
	"sync"
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
	wg          sync.WaitGroup
	ln          net.Listener
	connections map[int]net.Conn
	countConn   int
	log         Logger
	errLog      Logger
	isAlive     bool
}

func NewServer(port int, log Logger) *TCPServer {
	return &TCPServer{
		port: port,
		log:  log,
		errLog: func(msg string) {
			log("[Error]" + msg)
		},
	}
}

// Port gets the server's listening port
func (s TCPServer) Port() int {
	return s.port
}

// Start serves the TCP server and listens for connections
// A waitgroup needs have 1 for the TCP server and passed.
func (s *TCPServer) Start(wg sync.WaitGroup) (err error) {
	// Ensure caller's waitgroup is closed
	defer wg.Done()

	s.wg.Add(1)
	s.ln, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return
	}
	s.isAlive = true
	s.log(fmt.Sprintf("TCP Server started on port %d", s.port))

	// Close listener at end of function
	defer func() {
		s.ln.Close()
		s.wg.Done()
	}()

	// Handle each new connection
	for s.isAlive {
		s.wg.Add(1)
		// Block until new connection and accept it
		conn, err := s.ln.Accept()
		if err != nil {
			conn.Close() // Close connection
			s.errLog(fmt.Sprint("error accepting client connection:", err))
			continue // Proceed to block until next client connection
		}

		// Add connection to the slice
		s.connections[s.countConn] = conn
		connId := s.countConn
		s.countConn += 1

		// Handle each incoming packet
		for err != nil {
			var buf []byte
			_, err = conn.Read(buf)
			// Some handling here
		}

		// Packet handling loop is broken, clean up
		conn.Close()
		delete(s.connections, connId)
		s.wg.Done()
	}

	return
}

func (s *TCPServer) Stop() (err error) {
	// Close client connections
	for _, connection := range s.connections {
		connection.Close() // No error handling since we're trying to shut down anyway
		s.wg.Done()
	}

	// Close listener loop
	s.isAlive = false

	// Close listener
	err = s.ln.Close()

	// Block until server has been gracefully shut down
	s.wg.Wait()

	return
}
