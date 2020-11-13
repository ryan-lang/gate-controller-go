package messages

import (
	"gate/logical"
)

type VersionMessage struct {
	responseChan chan interface{}
	errChan      chan error
	address      int
}

type VersionResponse struct {
	Version string
}

func NewVersionMessage(addr int) *VersionMessage {
	return &VersionMessage{
		responseChan: make(chan interface{}, 1),
		errChan:      make(chan error, 3),
		address:      addr,
	}
}

func (m *VersionMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x56, nil, byte(m.address))
}

func (m *VersionMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x56
}

func (m *VersionMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- &VersionResponse{
		Version: string(p.Message),
	}
	return nil
}

func (m *VersionMessage) ResponseChan() chan interface{} {
	return m.responseChan
}

func (m *VersionMessage) ErrChan() chan error {
	return m.errChan
}
