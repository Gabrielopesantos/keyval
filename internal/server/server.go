package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"net"
	"sync"

	"github.com/gabrielopesantos/keyval/internal/command"
)

const (
	DEFAULT_READ_BUFFER_SIZE = 256
)

// Server is the entity that binds all components from the key value data store,
// listeners, storage interface, session interface, worker;
type Server struct {
	listener net.Listener // UDP isn't a listener;
	storage  sync.Map     // Will eventually be a storageManager
	logger   *slog.Logger
}

func New(addr string, logger *slog.Logger) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewServerFromListener(listener, logger), nil
}

func NewServerFromListener(listener net.Listener, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &Server{
		listener: listener,
		storage:  sync.Map{},
		logger:   logger,
	}
}

func (s *Server) AcceptConns() {
	defer s.listener.Close() // Returns err
	s.logger.Info("accepting connections on port 22122")

	// Listening loop
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error(fmt.Sprintf("failed to establish connection: %s", err))
		}

		go s.processConn(conn)
	}
}

func (s *Server) processConn(conn net.Conn) {
	defer conn.Close() // Returns err

	connReader := bufio.NewReader(conn)
	cmdLiteral, err := connReader.ReadString(' ')
	if err != nil {
		s.logger.Error(fmt.Sprintf("could not read command: %s", err))
		return
	}
	cmdLiteral = strings.Trim(cmdLiteral, "\n\r ")

	commandManager, err := command.NewCommand(cmdLiteral, s.logger)
	if err != nil {
		s.logger.Error(fmt.Sprintf("could not get command manager for command literal '%s': %s", err, cmdLiteral))
		// Write something back
		return
	}

	if command.HasAdditionalArguments(cmdLiteral) {
		err = commandManager.Parse(connReader)
		if err != nil {
			s.logger.Error(fmt.Sprintf("could not parse command: %s", err))
			// Write something back
			return
		}
	}

	// Might need some references; Worker or something;
	responseMessage := commandManager.Exec(&s.storage)

	s.logger.Debug(fmt.Sprintf("message to write: '%s'", responseMessage))
	n, err := conn.Write(responseMessage)
	if err != nil {
		s.logger.Error(fmt.Sprintf("could not write response: %s", err))
		return
	}
	for n != len(responseMessage) {
		n, err = conn.Write(responseMessage[n:])
		if err != nil {
			s.logger.Error(fmt.Sprintf("could not write response: %s", err))
			return
		}
	}
}
