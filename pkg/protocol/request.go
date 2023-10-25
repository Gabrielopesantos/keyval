package protocol

import (
	"errors"
	"fmt"
	"io"
)

type Request struct {
	Command   []byte
	Arguments [][]byte
}

func NewRequestFromReader(reader io.Reader) (*Request, error) {
	var command []byte
	itemsParsed, err := fmt.Fscanf(reader, "%s", &command)
	if err != nil {
		return nil, err
	}

	if itemsParsed == 0 {
		return nil, errors.New("no items parsed")
	}

	var arguments [][]byte
	for {
		var arg []byte
		_, err := fmt.Fscanf(reader, "%s", &arg)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		arguments = append(arguments, arg)
	}

	return &Request{Command: command, Arguments: arguments}, nil
}
