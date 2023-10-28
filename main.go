package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gabrielopesantos/keyval/internal/server"
)

var (
	listenPort   int
	loggingLevel int
)

func init() {
	flag.IntVar(&listenPort, "listen", 22122, "Port on which process listens")
	flag.IntVar(&loggingLevel, "logging-level", 0, "Server logging level (DEBUG: -4; INFO: 0; WARN: 4; ERROR: 8)")
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", listenPort)
	// TODO: Handler and writer should also be configurable
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(loggingLevel)}))
	srv, err := server.New(addr, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("could not start listening for connections on addr '%s': %s", addr, err))
	}

	srv.AcceptConns()
}
