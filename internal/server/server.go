package server

import (
	"bufio"
	"strings"

	// "bytes"

	"log"
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
}

func New(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewServerFromListener(listener), nil
}

func NewServerFromListener(listener net.Listener) *Server {
	return &Server{
		listener: listener,
		storage:  sync.Map{},
	}
}

func (s *Server) AcceptConns() {
	defer s.listener.Close() // Returns err
	log.Println("accepting connections on port 22122")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("failed to establish connection: %s", err)
		}

		go s.processConn(conn)
	}
}

func (s *Server) processConn(conn net.Conn) {
	defer conn.Close() // Returns err

	connReader := bufio.NewReader(conn)
	cmdLiteral, err := connReader.ReadString(' ')
	if err != nil {
		log.Printf("ERROR: could not read command: %s", err)
		return
	}
	cmdLiteral = strings.Trim(cmdLiteral, "\n\r ")

	commandManager, err := command.NewCommand(cmdLiteral)
	if err != nil {
		log.Printf("ERROR: could not get command manager for command literal '%s': %s", err, cmdLiteral)
		// Write something back
		return
	}

	if command.HasAdditionalArguments(cmdLiteral) {
		err = commandManager.Parse(connReader)
		if err != nil {
			log.Printf("ERROR: could not parse command: %s", err)
			// Write something back
			return
		}
	}

	// Might need some references; Worker or something;
	responseMessage := commandManager.Exec(&s.storage)

	log.Printf("INFO: Sending: '%s'", responseMessage)
	n, err := conn.Write(responseMessage)
	if err != nil {
		log.Printf("ERROR: could not write response: %s", err)
		return
	}
	for n != len(responseMessage) {
		n, err = conn.Write(responseMessage[n:])
		if err != nil {
			log.Printf("ERROR: could not write response: %s", err)
			return
		}
	}
}
