package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gabrielopesantos/keyval/internal/server"
)

var (
	listenPort int
)

func init() {
	flag.IntVar(&listenPort, "listen", 22122, "Port on which process listens")
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", listenPort)
	srv, err := server.New(addr)
	if err != nil {
		log.Fatalf("could not start listening for connections on addr '%s': %s", addr, err)
	}

	srv.AcceptConns()
}
