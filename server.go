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

type Connection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

// An Codec is classified as a function that can take in a slice of bytes and return the manipulated form of it
type Codec func([]byte)

// A Logger is classified as a function that can take in a string
type Logger func(string)

type Server struct {
	connections map[int]net.Conn
	isAlive     bool
	countConn   int
	port        int
	onPacket    func(Connection, []byte)
	onConnected func(Connection)
	encrypt     Codec
	decrypt     Codec
	errLog      Logger
	log         Logger
	ln          net.Listener
	wg          sync.WaitGroup
}

type ServerOption func(*Server)

func NewServer(options ...ServerOption) *Server {
	const (
		defaultPort = 8484
	)

	s := &Server{
		port:        defaultPort,
		isAlive:     false,
		connections: make(map[int]net.Conn),
	}

	for _, option := range options {
		option(s)
	}

	return s
}

// WithPort return a `ServerOption` which the Server constructor uses to modify its `port` member
func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

// WithLoggers returns a `ServerOption` which the Server constructor uses to modify its `logger` members
//
// If the `errLogger` parameter is left empty, then the errLogger function would use the `logger` parameter with [Error] prefixed.
func WithLoggers(logger Logger, errLogger Logger) ServerOption {
	return func(s *Server) {
		s.log = logger

		if errLogger == nil {
			s.errLog = func(msg string) {
				s.log(fmt.Sprint("[Error]", msg))
			}
		}
	}
}

// WithEncrypter returns a `ServerOption` which the Server constructor uses to modify its `encrypt` member
func WithEncrypter(encrypter Codec) ServerOption {
	return func(s *Server) {
		s.encrypt = encrypter
	}
}

// WithDecrypter returns a `ServerOption` which the Server constructor uses to modify its `decrypt` member
func WithDecrypter(decrypter Codec) ServerOption {
	return func(s *Server) {
		s.decrypt = decrypter
	}
}

// WithOnPacket returns a `ServerOption` which the Server constructor uses to modify its `onPacket` member
func WithOnPacket(onPacket func(Connection, []byte)) ServerOption {
	return func(s *Server) {
		s.onPacket = onPacket
	}
}

// WithOnConnected returns a `ServerOption` which the Server constructor uses to modify its `onConnected` member
func WithOnConnected(onConnected func(Connection)) ServerOption {
	return func(s *Server) {
		s.onConnected = onConnected
	}
}

// Port gets the server's listening port
func (s Server) Port() int {
	return s.port
}

// Start serves the TCP server and listens for connections
// A waitgroup needs have 1 for the TCP server and passed.
func (s *Server) Start(wg sync.WaitGroup) (err error) {
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
		s.onConnected(conn)
		s.log(fmt.Sprintf("New client connection made (ID: %d)", connId))

		// Handle each incoming packet
		for err == nil {
			// Read the packet without knowing its size
			buf := make([]byte, 2048) // We set the buffer to 2048 and shrink it later
			n, err := conn.Read(buf)  // Attempt to read from the connection
			if err != nil {
				// If cannot read the packet, end the loop and close connection
				s.errLog(fmt.Sprint("Could not read packet", err))
				break
			}

			data := buf[4:n]
			s.decrypt(data)        // Decrypt data if there is a decrypter
			s.onPacket(conn, data) // Send event to the outside
		}

		// Packet handling loop is broken, clean up
		conn.Close()
		delete(s.connections, connId)
		s.wg.Done()
	}

	return
}

func (s *Server) Stop() (err error) {
	// Close client connections
	for _, connection := range s.connections {
		connection.Close() // No error handling since we're trying to shut down anyway
		s.wg.Done()
	}

	s.isAlive = false  // Close listener loop
	err = s.ln.Close() // Close listener
	s.wg.Wait()        // Block until server has been gracefully shut down

	return
}
