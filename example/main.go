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

	port := tcpserve.WithPort(8484)
	loggers := tcpserve.WithLoggers(logger, nil)
	encrypter := tcpserve.WithEncrypter(encrypt)
	decrypter := tcpserve.WithDecrypter(nil) // You can omit this line
	onConnected := tcpserve.WithOnConnected(nil)
	onPacket := tcpserve.WithOnPacket(
		func(c tcpserve.Connection, b []byte) {
			log.Println(b)
		},
	) // Simple onPacket handler that just prints the bytes received

	wg.Add(1)
	server := tcpserve.NewServer(port, loggers, encrypter, decrypter, onConnected, onPacket)
	server.Start(wg)

	wg.Wait()
}
