package server

import (
	"bufio"
	"strings"

	// "bytes"
	"io"
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
	listener      net.Listener               // UDP isn't a listener;
	storage       sync.Map                   // Will eventually be a storageManager
	knownCommands map[string]command.Command // This would be in the worker or something;
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
		listener:      listener,
		storage:       sync.Map{},
		knownCommands: map[string]command.Command{"PING": &command.PingCommand{}},
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
	command, err := connReader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			log.Printf("INFO: connection closed before reading command: %s", err)
		}
		log.Printf("ERROR: unexpected error while reading message command: %s", err)
	}

	// Partial read until command is read; If not known, stop reading;
	command = strings.Trim(command, "\n\r ") // Is it a good idea to do this?
	log.Printf("INFO: '%s' command invoked", command)
	commandManager, ok := s.knownCommands[command]
	if !ok {
		log.Printf("ERROR: Unknown command, closing connection;") // Close conn?
	}

	// readBuffer := make([]byte, DEFAULT_READ_BUFFER_SIZE)
	if command == "PING" {
		goto noParameterCommands
	}
	// _, err = connReader.Read(readBuffer)
	// if err != nil {
	// 	log.Printf("ERROR: could not read incoming message: %s", err)
	// 	return
	// }

	// NOTE: For commands like `PING` that have nothing after command this isn't needed;
noParameterCommands:
	err = commandManager.Parse(connReader)
	if err != nil {
		log.Printf("ERROR: Could not parse command parameters;")
		return
	}
	// Might need some references; Worker or something;
	responseMessage := commandManager.Exec()

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
