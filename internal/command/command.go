package command

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/gabrielopesantos/keyval/internal/item"
	"github.com/gabrielopesantos/keyval/internal/storage"
)

const (
	PARAMETERS_DELIMITER = ' ' // NOTE: Not used
)

type Factory func(storageManager storage.Manager, logger *slog.Logger) Command

var commandFactories = map[string]Factory{
	"PING": NewPingCommand,
	"GET":  NewGetCommand,
	"ADD":  NewAddCommand,
}

func NewCommand(command string, storageManager storage.Manager, logger *slog.Logger) (Command, error) {
	factory, ok := commandFactories[command]
	if !ok {
		return nil, UnknownCommandError
	}
	return factory(storageManager, logger), nil
}

func HasAdditionalArguments(command string) bool {
	return command != "PING"
}

type Command interface {
	Exec() ([]byte, error)
	Parse(argsReader io.Reader) error
}

type BaseCommand struct {
	storageManager storage.Manager
	logger         *slog.Logger
}

type PingCommand struct {
	BaseCommand
}

// Also have a String method?

func NewPingCommand(storageManager storage.Manager, logger *slog.Logger) Command {
	return &PingCommand{BaseCommand{storageManager, logger}}
}

func (c *PingCommand) Parse(argsReader io.Reader) error {
	return nil
}

func (c *PingCommand) Exec() ([]byte, error) {
	return []byte("PONG\r\n"), nil
}

type GetCommand struct {
	BaseCommand
	Key string
}

func NewGetCommand(storageManager storage.Manager, logger *slog.Logger) Command {
	return &GetCommand{BaseCommand: BaseCommand{storageManager, logger}}
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

func (c *GetCommand) Exec() ([]byte, error) {
	c.logger.Debug(fmt.Sprintf("GET - Key: %s", c.Key))
	item, err := c.storageManager.Get(c.Key)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s\r\n", item.Value)), nil
}

type AddCommand struct {
	BaseCommand
	item *item.Item
}

func NewAddCommand(storageManager storage.Manager, logger *slog.Logger) Command {
	return &AddCommand{BaseCommand: BaseCommand{storageManager, logger}}
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

func (c *AddCommand) Exec() ([]byte, error) {
	c.logger.Debug(fmt.Sprintf("ADD - Key: %s; Value: %v", c.item.Key, c.item))
	if err := c.storageManager.Add(c.item); err != nil {
		return nil, err
	}
	return []byte("STORED\r\n"), nil
}
