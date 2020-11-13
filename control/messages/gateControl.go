package messages

import (
	"gate/logical"
)

type GateControlMessage struct {
	responseChan chan interface{}
	errChan      chan error
	request      *GateControlRequest
}

type GateControlRequest struct {
	Address                  int
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
		responseChan: make(chan interface{}, 1),
		errChan:      make(chan error, 3),
		request:      req,
	}
}

func (m *GateControlMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x43, m.request.Message(), byte(m.request.Address))
}

func (m *GateControlMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x43
}

func (m *GateControlMessage) HandleResponse(p *logical.Packet) error {
	// no data, just acknowledgment
	m.responseChan <- true
	return nil
}

func (m *GateControlMessage) ResponseChan() chan interface{} {
	return m.responseChan
}

func (m *GateControlMessage) ErrChan() chan error {
	return m.errChan
}

func (m *GateControlRequest) Message() []byte {
	var msg []byte

	if m.PushButtonOpen {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.PushButtonClose {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.PushButtonStop {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.OpenPartial {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.EmergencyOpen {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.EmergencyClose {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.OpenInterlock {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	if m.BlockExitVehicleDetector {
		msg = append(msg, 0x31)
	} else {
		msg = append(msg, 0x30)
	}

	return msg
}
