package messages

import (
	"gates/logical"
)

type VersionMessage struct {
}

type VersionResponse struct {
	Version string
}

func NewVersionMessage() *VersionMessage {
	return &VersionMessage{
		responseChan: make(chan *VersionResponse),
	}
}

func (m *VersionMessage) Packet() *logical.Packet {
	return &logical.Packet{
		MessageType: "V",
	}
}

func (m *VersionMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == "V"
}

func (m *VersionMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- &VersionResponse{
		Version: string(b.Message),
	}
	return nil
}

func (m *VersionMessage) ResponseChan() chan *VersionResponse {
	return m.responseChan
}
