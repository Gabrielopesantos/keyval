package main

import (
	"log"
	"net"

	"github.com/gabrielopesantos/keyval/internal/server"
)

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:22122")
	if err != nil {
		log.Fatalf("could not listening for connections on addr '127.0.0.1:22122': %s", err)
	}

	srv := server.NewServerFromListener(tcpListener)
	srv.AcceptConns()
}
