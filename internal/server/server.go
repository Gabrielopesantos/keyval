package server

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"sync"
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

func NewServerFromListener(listener net.Listener) *Server {
	return &Server{
		listener: listener,
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
	defer conn.Close() // Returns err // Defer doesn't work?
	readBuffer := make([]byte, DEFAULT_READ_BUFFER_SIZE)
	bytesRead, err := conn.Read(readBuffer)
	if err != nil {
		log.Printf("ERROR: could not read incoming message: %s", err)
		return
	}

	// Is this even possible?
	if bytesRead == 0 {
		log.Println("INFO: empty message...")
		return
	}

	command := s.parseMessage(readBuffer)
	if command == nil {
		return
	}

	responseMessage := command.Exec(s)
	log.Printf("INFO: Sending: '%s'", responseMessage)
	n, err := conn.Write(responseMessage)
	if err != nil {
		log.Printf("ERROR: could not write response: %s", err)
	}
	for n != len(responseMessage) {
		n, err = conn.Write(responseMessage[n:])
		if err != nil {
			log.Printf("ERROR: could not write response: %s", err)
		}
	}
}

func (s *Server) parseMessage(message []byte) Command {
	messageReader := bufio.NewReader(bytes.NewReader(message))
	command, err := messageReader.ReadSlice('\n')
	if err != nil {
		log.Printf("ERROR: could not read command from message: %s", err)
		return nil
	}
	if string(command) == "PING" {
		return &PingCommand{}
	}

	return &PingCommand{}
}

type Command interface {
	Exec(server *Server) []byte
}

type PingCommand struct{}

func (c *PingCommand) Exec(server *Server) []byte {
	return []byte("PONG")
}
