package messages

import (
	"gate/logical"
)

type ResetMessage struct {
}

func NewResetMessage() *ResetMessage {
	return &ResetMessage{}
}

func (m *ResetMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x52, nil, 254)
}
