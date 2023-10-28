package command

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/gabrielopesantos/keyval/internal/item"
)

const (
	PARAMETERS_DELIMITER = ' '
)

type CommandFactory func(*slog.Logger) Command

var commandFactories = map[string]CommandFactory{
	"PING": NewPingCommand,
	"GET":  NewGetCommand,
	"ADD":  NewAddCommand,
}

func NewCommand(command string, logger *slog.Logger) (Command, error) {
	factory, ok := commandFactories[command]
	if !ok {
		return nil, errors.New("ErrUnknownCommand") // ErrUnknownCommand
	}
	return factory(logger), nil
}

func HasAdditionalArguments(command string) bool {
	return command != "PING"
}

type Command interface {
	// "storage"
	Exec(storage *sync.Map) []byte
	Parse(argsReader io.Reader) error
}

type PingCommand struct {
	logger *slog.Logger
}

// Also be a String method?

func NewPingCommand(logger *slog.Logger) Command {
	return &PingCommand{logger: logger}
}

func (c *PingCommand) Parse(argsReader io.Reader) error {
	return nil
}

func (c *PingCommand) Exec(storage *sync.Map) []byte {
	return []byte("PONG\r\n")
}

type GetCommand struct {
	Key    string
	logger *slog.Logger
}

func NewGetCommand(logger *slog.Logger) Command {
	return &GetCommand{logger: logger}
}

func (c *GetCommand) Parse(argsReader io.Reader) error {
	format := "%s\r\n"
	var key string
	parsedItems, err := fmt.Fscanf(argsReader, format, &key)
	// NOTE: err != io.EOF
	if err != nil && err != io.EOF {
		return err
	}
	if parsedItems == 0 {
		return errors.New("ErrInvalidCommandArguments")
	}
	c.Key = key

	return nil
}

func (c *GetCommand) Exec(storage *sync.Map) []byte {
	c.logger.Debug(fmt.Sprintf("GET - Key: %s", c.Key))
	value, ok := storage.Load(c.Key)
	if !ok {
		return []byte("ERROR: Key not found\r\n")
	}

	itemValue, ok := value.(*item.Item)
	if !ok {
		// NOTE: Not expected to happen
		return []byte("ERROR: Could not parse stored key\r\n")
	}

	return []byte(fmt.Sprintf("%s\r\n", itemValue.Value))
}

type AddCommand struct {
	// Types
	Key    string
	Val    []byte
	Flags  uint8
	TTL    uint64
	logger *slog.Logger
}

func NewAddCommand(logger *slog.Logger) Command {
	return &AddCommand{logger: logger}
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
	c.logger.Debug(fmt.Sprintf("ADD - Key: %s; Value: %v", c.Key, item))
	storage.Store(c.Key, item)
	return []byte("STORED\r\n")
}
