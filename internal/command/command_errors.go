package command

import "errors"

var (
	UnknownCommandError          = errors.New("ErrUnknownCommand")
	InvalidCommandArgumentsError = errors.New("ErrInvalidCommandArguments")
)
