package server

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/gabrielopesantos/keyval/internal/command"
	"github.com/gabrielopesantos/keyval/internal/storage"
	"net"
)

const (
	DEFAULT_READ_BUFFER_SIZE = 256
)

// Server is the entity that binds all components from the key value data store,
// listeners, storage interface, session interface, worker;
type Server struct {
	listener       net.Listener // UDP isn't a listener;
	storageManager storage.Manager
	logger         *slog.Logger
}

func New(addr string, storageManager storage.Manager, logger *slog.Logger) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewServerFromListener(listener, storageManager, logger), nil
}

func NewServerFromListener(listener net.Listener, storageManager storage.Manager, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &Server{
		listener:       listener,
		storageManager: storageManager,
		logger:         logger,
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

	// Parse and execute command
	connReader := bufio.NewReader(conn)
	respMessage, err := s.parseAndExecuteCommand(connReader)
	s.logger.Debug(fmt.Sprintf("response message: '%s'", respMessage))
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to interpret message: %s", err))
	}

	n, err := conn.Write(respMessage)
	if err != nil {
		s.logger.Error(fmt.Sprintf("could not write response: %s", err))
		return
	}
	for n != len(respMessage) {
		n, err = conn.Write(respMessage[n:])
		if err != nil {
			s.logger.Error(fmt.Sprintf("could not write response: %s", err))
			return
		}
	}
}

func (s *Server) parseAndExecuteCommand(connReader *bufio.Reader) ([]byte, error) {
	cmdLiteral, err := connReader.ReadString(' ')
	if err != nil && err != io.EOF {
		return command.READ_COMMAND_ERROR_RESP, fmt.Errorf("could not read message command: %s", err)
	}
	cmdLiteral = strings.Trim(cmdLiteral, "\n\r ")

	commandExecutor := command.NewCommand(cmdLiteral, s.storageManager, s.logger)
	if commandExecutor == nil {
		return command.UNKNOWN_COMMAND_ERROR_RESP, fmt.Errorf("unknown command provided, '%s'", cmdLiteral)
	}

	if command.HasArguments(cmdLiteral) {
		err = commandExecutor.Parse(connReader)
		if err != nil {
			return command.PARSE_COMMAND_ERROR_RESP, fmt.Errorf("could not parse command arguments: %s", err)
		}
	}

	respMessage, err := commandExecutor.Exec()
	if err != nil {
		return respMessage, fmt.Errorf("could not execute requested command: %s", err)
	}

	return respMessage, nil
}
