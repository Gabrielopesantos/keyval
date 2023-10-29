package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gabrielopesantos/keyval/internal/server"
	"github.com/gabrielopesantos/keyval/internal/storage"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.Level(loggingLevel)}))
	addr := fmt.Sprintf("127.0.0.1:%d", listenPort)
	storageManager := storage.NewSyncMapStorage(false, logger)
	// TODO: Handler and writer should also be configurable
	srv, err := server.New(addr, storageManager, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("could not start listening for connections on addr '%s': %s", addr, err))
	}

	srv.AcceptConns()
}
