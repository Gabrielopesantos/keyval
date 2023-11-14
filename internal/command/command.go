package command

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/gabrielopesantos/keyval/internal/item"
	"github.com/gabrielopesantos/keyval/internal/storage"
)

// NOTE: Maybe these shouldn't be here?
var (
	PARAMETERS_DELIMITER                 = ' '
	CR_LF                                = []byte("\r\n")
	PONG_RESP                            = []byte("PONG\r\n")
	OK_RESP                              = []byte("Ok\r\n")
	READ_COMMAND_ERROR_RESP              = []byte("ErrReadCommand\r\n")
	UNKNOWN_COMMAND_ERROR_RESP           = []byte("ErrUnknownCommand\r\n")
	PARSE_COMMAND_ERROR_RESP             = []byte("ErrParseCommand\r\n")
	INVALID_COMMNAD_ARGUMENTS_ERROR_RESP = []byte("ErrInvalidCommandArguments\r\n")
	KEY_EXISTS_ERROR_RESP                = []byte("ErrKeyExists\r\n")
	KEY_NOT_EXISTS_ERROR_RESP            = []byte("ErrKeyNotExists\r\n")
)

type Factory func(storageManager storage.Manager, logger *slog.Logger) Command

var commandFactories = map[string]Factory{
	"PING": NewPingCommand,
	"GET":  NewGetCommand,
	"ADD":  NewAddCommand,
	"DEL":  NewDeleteCommand,
}

func NewCommand(command string, storageManager storage.Manager, logger *slog.Logger) Command {
	factory, ok := commandFactories[command]
	if !ok {
		return nil
	}
	return factory(storageManager, logger)
}

func HasArguments(command string) bool {
	return command != "PING"
}

type Command interface {
	Parse(argsReader io.Reader) error
	Exec() ([]byte, error)
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
	return PONG_RESP, nil
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
		return fmt.Errorf("number of parsed arguments (%d) differs from the expected, 1", parsedItems)
	}
	c.Key = key

	return nil
}

func (c *GetCommand) Exec() ([]byte, error) {
	c.logger.Debug(fmt.Sprintf("GET - Key: %s", c.Key))
	item, err := c.storageManager.Get(c.Key)
	if err != nil {
		// TODO: Can also be a different error
		return KEY_NOT_EXISTS_ERROR_RESP, err
	}
	// TODO
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
		return fmt.Errorf("number of parsed items (%d) differs from the expected, %d", parsedItems, expectedArguments)
	}
	c.item = item

	return nil
}

func (c *AddCommand) Exec() ([]byte, error) {
	c.logger.Debug(fmt.Sprintf("ADD - Key: %s; Value: %s; Flags: %d; TTL: %d", c.item.Key, c.item.Value, c.item.Flags, c.item.TTL))
	if err := c.storageManager.Add(c.item); err != nil {
		return KEY_EXISTS_ERROR_RESP, err
	}
	// TODO
	return []byte("STORED\r\n"), nil
}

type DeleteCommand struct {
	BaseCommand
	Key string
}

// NOTE: Reconsider these, New*Command,  functions
func NewDeleteCommand(storageManager storage.Manager, logger *slog.Logger) Command {
	return &DeleteCommand{BaseCommand: BaseCommand{storageManager, logger}}
}

func (c *DeleteCommand) Parse(argsReader io.Reader) error {
	format := "%s\r\n"
	var key string
	parsedItems, err := fmt.Fscanf(argsReader, format, &key)
	// NOTE: err != io.EOF
	if err != nil && err != io.EOF {
		return err
	}
	if parsedItems == 0 {
		return fmt.Errorf("number of parsed arguments (%d) differs from the expected, 1", parsedItems)
	}
	c.Key = key

	return nil
}

func (c *DeleteCommand) Exec() ([]byte, error) {
	c.logger.Debug(fmt.Sprintf("DEL - Key: %s", c.Key))
	c.storageManager.Delete(c.Key)
	return []byte("OK\r\n"), nil
}
