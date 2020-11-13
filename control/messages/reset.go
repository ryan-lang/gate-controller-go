package messages

import (
	"gate/logical"
)

type ResetMessage struct {
	address int
	errChan chan error
}

func NewResetMessage(addr int) *ResetMessage {
	return &ResetMessage{
		address: addr,
		errChan: make(chan error, 3),
	}
}

func (m *ResetMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x52, nil, byte(m.address))
}

func (m *ResetMessage) ErrChan() chan error {
	return m.errChan
}
