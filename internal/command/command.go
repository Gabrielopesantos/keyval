package command

import (
	"bufio"
	"errors"
	"log"
	"strconv"
)

const (
	PARAMETERS_DELIMITER = ' '
)

type Command interface {
	// Exec(server *server.Server) []byte
	Exec() []byte
	// Parse(commandParameters []byte) error
	Parse(connReader *bufio.Reader) error
}

type PingCommand struct{}

// Also be a String method?

// func NewPingCommand() Command {
// 	return &PingCommand{}
// }

func (c *PingCommand) Parse(connReader *bufio.Reader) error {
	return nil
}

// func (c *PingCommand) Exec(server *server.Server) []byte {
func (c *PingCommand) Exec() []byte {
	return []byte("PONG")
}

type AddCommand struct {
	// Types
	Key   string
	Val   []byte
	Flags uint8
	TTL   uint64
}

func (c *AddCommand) Parse(connReader *bufio.Reader) error {
	// Read key
	key, err := connReader.ReadString(PARAMETERS_DELIMITER)
	if err != nil {
		log.Printf("ERROR: 'ADD' command: could not read item key parameter: %s", err)
		return errors.New("INVALID")
	}
	c.Key = key

	// Read flags
	// Would only be a byte
	flagsStr, err := connReader.ReadString(PARAMETERS_DELIMITER)
	if err != nil {
		log.Printf("ERROR: 'ADD' command: could not read item flags parameter: %s", err)
		return errors.New("INVALID")
	}
	flags, err := strconv.ParseUint(flagsStr, 10, 8)
	if err != nil {
		log.Printf("ERROR: 'ADD' command: could not parse item flags into an integer: %s", err)
		return errors.New("INVALID")
	}
	c.Flags = uint8(flags)

	// Read TTL
	ttlStr, err := connReader.ReadString(PARAMETERS_DELIMITER)
	if err != nil {
		log.Printf("ERROR: 'ADD' command: could not read item ttl parameter: %s", err)
		return errors.New("INVALID")
	}
	ttl, err := strconv.ParseUint(ttlStr, 10, 64)
	if err != nil {
		log.Printf("ERROR: 'ADD' command: could not parse item ttl into an integer: %s", err)
		return errors.New("INVALID")
	}
	c.TTL = ttl
	return nil
}

func (c *AddCommand) Exec() []byte {
	return nil
}
