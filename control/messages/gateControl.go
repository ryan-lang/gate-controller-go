package messages

import (
	"gate/logical"
)

type GateControlMessage struct {
	responseChan chan interface{}
}

type GateControlRequest struct {
	PushButtonOpen           bool
	PushButtonClose          bool
	PushButtonStop           bool
	OpenPartial              bool
	EmergencyOpen            bool
	EmergencyClose           bool // latched
	OpenInterlock            bool // latched
	BlockExitVehicleDetector bool // latched
}

func NewGateControlMessage(req *GateControlRequest) *GateControlMessage {
	return &GateControlMessage{
		responseChan: make(chan interface{}),
	}
}

func (m *GateControlMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x43, nil, 254)
}

func (m *GateControlMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x43
}

func (m *GateControlMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- true
	return nil
}

func (m *GateControlMessage) ResponseChan() chan interface{} {
	return m.responseChan
}
