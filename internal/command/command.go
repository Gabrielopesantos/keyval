package command

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/gabrielopesantos/keyval/internal/item"
)

type CommandFactory func() Command

var commandFactories = map[string]CommandFactory{
	"PING": NewPingCommand,
	"ADD":  NewAddCommand,
}

const (
	PARAMETERS_DELIMITER = ' '
)

func NewCommand(command string) (Command, error) {
	factory, ok := commandFactories[command]
	if !ok {
		return nil, errors.New("ErrUnknownCommand") // ErrUnknownCommand
	}
	return factory(), nil
}

type Command interface {
	// "storage"
	Exec(storage *sync.Map) []byte
	Parse(argsReader io.Reader) error
}

type PingCommand struct{}

// Also be a String method?

func NewPingCommand() Command {
	return &PingCommand{}
}

func (c *PingCommand) Parse(argsReader io.Reader) error {
	return nil
}

func (c *PingCommand) Exec(storage *sync.Map) []byte {
	return []byte("PONG")
}

type AddCommand struct {
	// Types
	Key   string
	Val   []byte
	Flags uint8
	TTL   uint64
}

func NewAddCommand() Command {
	return &AddCommand{}
}

func (c *AddCommand) Parse(argsReader io.Reader) error {
	// key flags ttl val_size value
	expectedArguments := 5
	format := "%s %d %d %d\r\n%s\r\n"
	var valSize uint
	// Consider reading "header" first and read value from chars in valSize
	arguments := []interface{}{&c.Key, &c.Flags, &c.TTL, &valSize, &c.Val}
	parsedItems, err := fmt.Fscanf(argsReader, format, arguments...)
	if err != nil && err != io.EOF {
		return err
	}
	if parsedItems != expectedArguments {
		return errors.New("ErrInvalidCommandArguments")
	}

	return nil
}

func (c *AddCommand) Exec(storage *sync.Map) []byte {
	item := item.New(c.Key, c.Val, c.Flags, c.TTL)
	fmt.Printf("About to store: %+v", item)
	storage.Store(c.Key, &item)
	return []byte("STORED")
}

func HasAdditionalArguments(command string) bool {
	return command != "PING"
}
