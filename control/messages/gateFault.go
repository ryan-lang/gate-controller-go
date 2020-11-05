package messages

import (
	"gate/logical"
)

type GateFaultMessage struct {
	responseChan chan interface{}
}

type GateFaultResponse struct {
	NumberOfFaults int16
	FaultCodes     []int16
}

func NewGateFaultMessage() *GateFaultMessage {
	return &GateFaultMessage{
		responseChan: make(chan interface{}),
	}
}

func (m *GateFaultMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x46, nil, 254)
}

func (m *GateFaultMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x46
}

func (m *GateFaultMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- &GateFaultResponse{}
	return nil
}

func (m *GateFaultMessage) ResponseChan() chan interface{} {
	return m.responseChan
}
