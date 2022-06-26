package tcpserve

import (
	"fmt"
	"net"
	"sync"
)

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
	// Default options
	const (
		defaultPort = 8484
	)

	// Create Server object
	s := &Server{
		port:        defaultPort,
		isAlive:     false,
		connections: make(map[int]net.Conn),
	}

	// Call each option
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
	// Ensure caller's wait group is decremented when listener is closed
	defer wg.Done()

	s.wg.Add(1) // Increment wait group for the listener
	s.ln, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.wg.Done() // Decrement wait group for the listener
		return      // Return with error
	}
	// Listener server is alive
	s.isAlive = true
	s.log(fmt.Sprintf("TCP Server started on port %d", s.port))

	// Ensure listener is closed at end of function
	defer func() {
		s.ln.Close() // Close listener server
		s.wg.Done()  // Decrement wait group for listener
	}()

	// Handle each new connection
	for s.isAlive {
		s.wg.Add(1)                // Increment waitgroup for this connection
		conn, err := s.ln.Accept() // Block until new connection and accept it
		if err != nil {
			s.errLog(fmt.Sprint("error accepting client connection:", err))
			conn.Close() // Close connection
			s.wg.Done()  // Decrement wait group for connection
			continue     // Proceed to block until next client connection
		}

		go s.handleConn(conn)
	}

	return
}

// handleConn listens for new packets
func (s *Server) handleConn(conn net.Conn) {
	// Add connection to the slice
	id := s.countConn        // Set the current connection's ID
	s.connections[id] = conn // Add connection to the connections map with key = id
	s.countConn += 1         // Increment connection count for next ID
	s.onConnected(conn)      // Send onConnected to the outside
	s.log(fmt.Sprintf("New client connection made (ID: %d)", id))

	// Handle each incoming packet
	for {
		// Read the packet without knowing its size
		buf := make([]byte, 2048) // We set the buffer to 2048 and shrink it later
		n, err := conn.Read(buf)  // Attempt to read from the connection
		if err != nil {
			// If cannot read the packet, end the loop and close connection
			s.errLog(fmt.Sprintf("Closing connection (ID: %d). Could not read packet: %s", id, err))
			break
		}

		data := buf[4:n]       // Make a new byte slice from buffer containing the correct size packet
		s.decrypt(data)        // Decrypt data if there is a decrypter
		s.onPacket(conn, data) // Send event to the outside
	}

	// Packet handling loop is broken, clean up
	conn.Close()              // Close connection
	delete(s.connections, id) // Remove connection from connections map
	s.wg.Done()               // Decrement wait group for listener
}

// WriteToId sends the byte slice to the specified connection `id`
func (s *Server) WriteToId(message []byte, id int) {
	if connection, ok := s.connections[id]; ok {
		connection.Write(message)
	}
}

// WriteToAll sends the byte slice to all open connections
func (s *Server) WriteToAll(message []byte) {
	for _, connection := range s.connections {
		connection.Write(message)
	}
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
