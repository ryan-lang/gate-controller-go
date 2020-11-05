package messages

import (
	"gate/logical"
)

type VersionMessage struct {
	responseChan chan interface{}
}

type VersionResponse struct {
	Version string
}

func NewVersionMessage() *VersionMessage {
	return &VersionMessage{
		responseChan: make(chan interface{}),
	}
}

func (m *VersionMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x56, nil, 0x01)
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
