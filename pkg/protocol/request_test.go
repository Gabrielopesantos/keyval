package protocol

import (
	"bytes"
	"testing"
)

func TestRequestFromReader(t *testing.T) {
	addCommand := []byte("add color 0 3600 3 red")
	commandReader := bytes.NewBuffer(addCommand)
	request, err := NewRequestFromReader(commandReader)
	if err != nil {
		t.Fatalf("NewRequestFromReader failed: %v", err)
	}
	if !bytes.Equal(request.Command, []byte("add")) {
		t.Errorf("request.Command = %q, want %q", request.Command, []byte("add"))
	}
	if len(request.Arguments) != 5 {
		t.Errorf("len(request.Arguments) = %d, want %d", len(request.Arguments), 5)
	}
}
