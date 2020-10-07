package messages

import (
	"gates/logical"
)

type GateControlMessage struct {
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

func NewGateControlMessage() *GateControlMessage {
	return &GateControlMessage{
		responseChan: make(chan bool),
	}
}

func (m *GateControlMessage) Packet() *logical.Packet {
	return &logical.Packet{
		MessageType: "C",
	}
}

func (m *GateControlMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == "C"
}

func (m *GateControlMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- true
	return nil
}

func (m *GateControlMessage) ResponseChan() chan bool {
	return m.responseChan
}
