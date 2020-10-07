package messages

import (
	"gates/logical"
)

type GateStatusMessage struct {
}

type GateStatusResponse struct {
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

func NewGateStatusMessage() *GateStatusMessage {
	return &GateStatusMessage{
		responseChan: make(chan *GateStatusResponse),
	}
}

func (m *GateStatusMessage) Packet() *logical.Packet {
	return &logical.Packet{
		MessageType: "S",
	}
}

func (m *GateStatusMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == "S"
}

func (m *GateStatusMessage) HandleResponse(p *logical.Packet) error {
	m.responseChan <- &GateStatusResponse{
		MessageID: p.Message[0],
	}
	return nil
}

func (m *GateStatusMessage) ResponseChan() chan bool {
	return m.responseChan
}

const (
	STATUS_RESET              = 0
	STATUS_OPEN_INPROGRESS    = 1
	STATUS_OPEN_COMPLETE      = 2
	STATUS_CLOSE_INPROGRESS   = 3
	STATUS_CLOSE_COMPLETE     = 4
	STATUS_STOPPED            = 5
	STATUS_GEB                = 6
	STATUS_IES                = 7
	STATUS_ELD                = 8
	STATUS_SLD_HLD            = 9
	STATUS_IOLD               = 10
	STATUS_OOLD               = 11
	STATUS_PEO                = 12
	STATUS_PEC                = 13
	STATUS_OI                 = 14
	STATUS_LI                 = 15
	STATUS_POWER_LOCK         = 16
	STATUS_MODE4_IOLD_OOLD    = 17
	STATUS_ALERT14            = 18
	STATUS_OPEN_CMD           = 19
	STATUS_ENTRAPMENT         = 20
	STATUS_RELEARN_MODE       = 21
	STATUS_FAULT              = 22
	STATUS_ERROR              = 23
	STATUS_ALERT              = 24
	STATUS_EMOPEN_INPROGRESS  = 25
	STATUS_EMOPEN_COMPLETE    = 26
	STATUS_EMCLOSE_INPROGRESS = 27
	STATUS_EMCLOSE_COMPLETE   = 28
	STATUS_PEB                = 29
	STATUS_GEC                = 30
	STATUS_GEO                = 31
)

const (
	RESET_STATE            = 0
	LEARNLIMITSTOP_STATE   = 1
	LEARNLIMITOPEN_STATE   = 2
	LEARNLIMITCLOSE_STATE  = 3
	NORMALSTOP_STATE       = 4
	CHECKOPEN_STATE        = 5
	PEP2OPEN_STATE         = 6
	WARNB4OPEN             = 7
	NORMALOPEN_STATE       = 8
	REVERSE2CLOSEPEO_STATE = 9
	WAITPEO_STATE          = 10
	DELAYPEO_STATE         = 11
	CHECKPECLOSE_STATE     = 12
	PEP2CLOSE_STATE        = 13
	WARNB4CLOSE_STATE      = 14
	NORMALCLOSE_STATE      = 15
	WAITVD_STATE           = 16
	REVERSE2OPENPEC_STATE  = 17
	WAITPE_STATE           = 18
)
