package command

import (
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/gabrielopesantos/keyval/internal/item"
)

const (
	PARAMETERS_DELIMITER = ' ' // NOTE: Not used
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
		return nil, UnknownCommandError
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

// Also have a String method?

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
		return InvalidCommandArgumentsError
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
		c.logger.Warn("Unexpected condition evaluated:") // TODO:
		return []byte("ERROR: Could not parse stored key\r\n")
	}

	return []byte(fmt.Sprintf("%s\r\n", itemValue.Value))
}

type AddCommand struct {
	item   *item.Item
	logger *slog.Logger
}

func NewAddCommand(logger *slog.Logger) Command {
	return &AddCommand{logger: logger}
}

func (c *AddCommand) Parse(argsReader io.Reader) error {
	item := item.InitItem()
	// key flags ttl val_size value
	expectedArguments := 5
	format := "%s %d %d %d\r\n%s\r\n"
	var valSize uint
	// Consider reading "header" first and read value from chars in valSize
	arguments := []interface{}{&item.Key, &item.Flags, &item.TTL, &valSize, &item.Value}
	parsedItems, err := fmt.Fscanf(argsReader, format, arguments...)
	if err != nil && err != io.EOF {
		return err
	}
	if parsedItems != expectedArguments {
		return InvalidCommandArgumentsError
	}
	c.item = item

	return nil
}

func (c *AddCommand) Exec(storage *sync.Map) []byte {
	c.logger.Debug(fmt.Sprintf("ADD - Key: %s; Value: %v", c.item.Key, c.item))
	storage.Store(c.item.Key, c.item)
	return []byte("STORED\r\n")
}
