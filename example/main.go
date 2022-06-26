package main

import (
	"log"
	"sync"

	"github.com/matthieutran/tcpserve"
)

func main() {
	var wg sync.WaitGroup

	logger := func(msg string) {
		log.Println(msg)
	}

	encrypt := func(b []byte) {
		// do something with bytes
	}

	decrypt := func(b []byte) {
		// do something with bytes
	}

	handshake := func(s *tcpserve.Session) {
		s.SetEncrypter(encrypt)
		s.SetDecrypter(decrypt)

		s.Write([]byte{14, 0, 83, 0, 1, 0, 49, 87, 227, 158, 226, 254, 18, 15, 233, 8})
	}

	port := tcpserve.WithPort(8484)
	loggers := tcpserve.WithLoggers(logger, nil)
	onConnected := tcpserve.WithOnConnected(handshake)
	onPacket := tcpserve.WithOnPacket(
		func(s *tcpserve.Session, b []byte) {
			log.Println("Packet:", b)
		},
	) // Simple onPacket handler that just prints the bytes received

	wg.Add(1)
	server := tcpserve.NewServer(port, loggers, onConnected, onPacket)
	server.Start(wg)

	wg.Wait()
}
