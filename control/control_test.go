package control_test

import (
	"bytes"
	"gate/control"
	//"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func TestReset(t *testing.T) {
	logical := &mockLogical{}
	c := control.New(logical)

	go c.Start()

	return
}

type mockLogical struct {
	packet *logical.Packet
}

func (s *mockLogical) Read() (packet *Packet, err error) {
	return s.packet, nil
}

func (s *mockLogical) Write([]byte) (int, error) {
	return 0, nil
}
