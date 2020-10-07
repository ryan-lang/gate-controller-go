package messages

import (
	"gates/logical"
)

type GateFaultMessage struct {
}

type GateFaultResponse struct {
	MessageID             byte // new status
	LastCommandStatus     byte
	CurrentOperatorState  byte
	FaultsPresent         byte
	BatteryState          byte
	ACPresent             byte
	OpenLimit             byte
	CloseLimit            byte
	PartialOpenLimit      byte
	ExitLoop              byte
	InnerObstructionLoop  byte
	OuterObstructionLoop  byte
	ResetShadowLoop       byte
	Relays                []byte
	PhotoEyeOpen          byte
	PhotoEyeClose         byte
	GateEdgeBoth          byte
	GateEdgeClose         byte
	GateEdgeOpen          byte
	PhotoEyeBoth          byte
	OpenTooLong           byte
	Tailgater             byte
	Loitering             byte
	TransientVehicleCount int
	TenantVehicleCount    int
	SpecialVehicleCount   int
	UnknownVehicleCount   int
	CycleCount            int
	ELDCount              int
	IOLDCount             int
	OOLDCount             int
	HLDCLDCount           int
	TransientVends        int
	TenantVends           int
	SpecialVends          int
	ManualVends           int
}

func NewGateFaultMessage() *GateFaultMessage {
	return &GateFaultMessage{
		responseChan: make(chan bool),
	}
}

func (m *GateFaultMessage) Packet() *logical.Packet {
	return &logical.Packet{
		MessageType: "C",
	}
}

func (m *GateFaultMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == "C"
}

func (m *GateFaultMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- &GateFaultResponse{}
	return nil
}

func (m *GateFaultMessage) ResponseChan() chan bool {
	return m.responseChan
}
