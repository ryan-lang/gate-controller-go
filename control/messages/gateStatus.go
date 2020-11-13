package messages

import (
	"encoding/json"
	"gate/logical"
	"strconv"
)

type GateStatusMessage struct {
	responseChan chan interface{}
	errChan      chan error
	address      int
}

type GateStatusResponse struct {
	LastCommandStatus     int16
	CurrentOperatorState  int16
	FaultsPresent         bool
	BatteryState          int16
	ACPresent             bool
	OpenLimit             bool
	CloseLimit            bool
	PartialOpenLimit      bool
	ExitLoop              bool
	InnerObstructionLoop  bool
	OuterObstructionLoop  bool
	ResetShadowLoop       bool
	Relay1                byte
	Relay2                byte
	Relay3                byte
	PhotoEyeOpen          bool
	PhotoEyeClose         bool
	GateEdgeBoth          bool
	GateEdgeClose         bool
	GateEdgeOpen          bool
	PhotoEyeBoth          bool
	OpenTooLong           bool
	Tailgater             bool
	Loitering             bool
	TransientVehicleCount int16
	TenantVehicleCount    int16
	SpecialVehicleCount   int16
	UnknownVehicleCount   int16
	CycleCount            int16
	ELDCount              int16
	IOLDCount             int16
	OOLDCount             int16
	HLDCLDCount           int16
	TransientVends        int16
	TenantVends           int16
	SpecialVends          int16
	ManualVends           int16
}

func NewGateStatusMessage(addr int) *GateStatusMessage {
	return &GateStatusMessage{
		responseChan: make(chan interface{}, 1),
		errChan:      make(chan error, 3),
		address:      addr,
	}
}

func (m *GateStatusMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x53, nil, byte(m.address))
}

func (m *GateStatusMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x4e
}

func (m *GateStatusResponse) Diff(s *GateStatusResponse) bool {
	return m.LastCommandStatus != s.LastCommandStatus ||
		m.CurrentOperatorState != s.CurrentOperatorState ||
		m.FaultsPresent != s.FaultsPresent ||
		m.BatteryState != s.BatteryState ||
		m.ACPresent != s.ACPresent ||
		m.OpenLimit != s.OpenLimit ||
		m.CloseLimit != s.CloseLimit ||
		m.PartialOpenLimit != s.PartialOpenLimit ||
		m.ExitLoop != s.ExitLoop ||
		m.InnerObstructionLoop != s.InnerObstructionLoop ||
		m.OuterObstructionLoop != s.OuterObstructionLoop ||
		m.ResetShadowLoop != s.ResetShadowLoop ||
		m.Relay1 != s.Relay1 ||
		m.Relay2 != s.Relay2 ||
		m.Relay3 != s.Relay3 ||
		m.PhotoEyeOpen != s.PhotoEyeOpen ||
		m.PhotoEyeClose != s.PhotoEyeClose ||
		m.GateEdgeOpen != s.GateEdgeOpen ||
		m.GateEdgeClose != s.GateEdgeClose ||
		m.GateEdgeBoth != s.GateEdgeBoth ||
		m.PhotoEyeBoth != s.PhotoEyeBoth ||
		m.OpenTooLong != s.OpenTooLong ||
		m.Tailgater != s.Tailgater ||
		m.Loitering != s.Loitering ||
		m.TransientVehicleCount != s.TransientVehicleCount ||
		m.TenantVehicleCount != s.TenantVehicleCount ||
		m.SpecialVehicleCount != s.SpecialVehicleCount ||
		m.UnknownVehicleCount != s.UnknownVehicleCount ||
		m.CycleCount != s.CycleCount ||
		m.ELDCount != s.ELDCount ||
		m.IOLDCount != s.IOLDCount ||
		m.OOLDCount != s.OOLDCount ||
		m.HLDCLDCount != s.HLDCLDCount ||
		m.TransientVends != s.TransientVends ||
		m.TenantVends != s.TenantVends ||
		m.SpecialVends != s.SpecialVends ||
		m.ManualVends != s.ManualVends
}

func (m *GateStatusMessage) HandleResponse(p *logical.Packet) error {
	if len(p.Message) == 0 {
		return &ErrInvalidResponse{}
	}

	r := &GateStatusResponse{}

	{
		c, _ := strconv.Atoi(string(p.Message[0:2]))
		r.LastCommandStatus = int16(c)
	}

	{
		c, _ := strconv.Atoi(string(p.Message[2:4]))
		r.CurrentOperatorState = int16(c)
	}

	{
		if string(p.Message[4]) == "1" {
			r.FaultsPresent = true
		}
	}

	{
		c, _ := strconv.Atoi(string(p.Message[5]))
		r.BatteryState = int16(c)
	}

	{
		if string(p.Message[6]) == "1" {
			r.ACPresent = true
		}
	}

	{
		if string(p.Message[7]) == "1" {
			r.OpenLimit = true
		}
	}

	{
		if string(p.Message[8]) == "1" {
			r.CloseLimit = true
		}
	}

	{
		if string(p.Message[9]) == "1" {
			r.PartialOpenLimit = true
		}
	}

	{
		if string(p.Message[10]) == "1" {
			r.ExitLoop = true
		}
	}

	{
		if string(p.Message[11]) == "1" {
			r.InnerObstructionLoop = true
		}
	}

	{
		if string(p.Message[12]) == "1" {
			r.OuterObstructionLoop = true
		}
	}

	{
		if string(p.Message[13]) == "1" {
			r.ResetShadowLoop = true
		}
	}

	{
		r.Relay1 = p.Message[14]
		r.Relay2 = p.Message[15]
		r.Relay3 = p.Message[16]
		// fmt.Printf("relay1: % x/%s\n", r.Relay1, string(r.Relay1))
		// fmt.Printf("relay2: % x/%s\n", r.Relay2, string(r.Relay2))
		// fmt.Printf("relay3: % x/%s\n", r.Relay3, string(r.Relay3))
	}

	{
		if string(p.Message[17]) == "1" {
			r.PhotoEyeOpen = true
		}
	}

	{
		if string(p.Message[18]) == "1" {
			r.PhotoEyeClose = true
		}
	}

	{
		if string(p.Message[19]) == "1" {
			r.GateEdgeBoth = true
		}
	}

	{
		if string(p.Message[20]) == "1" {
			r.GateEdgeClose = true
		}
	}

	{
		if string(p.Message[21]) == "1" {
			r.GateEdgeOpen = true
		}
	}

	{
		if string(p.Message[22]) == "1" {
			r.PhotoEyeBoth = true
		}
	}

	// bytes 23-27 reserved
	{
		if string(p.Message[28]) == "1" {
			r.OpenTooLong = true
		}
	}

	{
		if string(p.Message[29]) == "1" {
			r.Tailgater = true
		}
	}

	{
		if string(p.Message[30]) == "1" {
			r.Loitering = true
		}
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[31:37]), 16, 16)
		r.TransientVehicleCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[37:43]), 16, 16)
		r.TenantVehicleCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[43:49]), 16, 16)
		r.SpecialVehicleCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[49:55]), 16, 16)
		r.UnknownVehicleCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[55:61]), 16, 16)
		r.CycleCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[61:65]), 16, 16)
		r.ELDCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[65:69]), 16, 16)
		r.IOLDCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[69:73]), 16, 16)
		r.OOLDCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[73:77]), 16, 16)
		r.HLDCLDCount = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[77:81]), 16, 16)
		r.TransientVends = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[81:85]), 16, 16)
		r.TenantVends = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[85:89]), 16, 16)
		r.SpecialVends = int16(i)
	}

	{
		i, _ := strconv.ParseInt(string(p.Message[89:93]), 16, 16)
		r.ManualVends = int16(i)
	}

	m.responseChan <- r
	return nil
}

func (m *GateStatusMessage) ResponseChan() chan interface{} {
	return m.responseChan
}

func (m *GateStatusMessage) ErrChan() chan error {
	return m.errChan
}

func (m *GateStatusResponse) String() string {
	b, _ := json.MarshalIndent(m, "", " ")
	return string(b)
}

// command status (3.2.2.1)
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

// operator states (3.2.2.2)
const (
	RESET_STATE            = 0
	LEARNLIMITSTOP_STATE   = 1
	LEARNLIMITOPEN_STATE   = 2
	LEARNLIMITCLOSE_STATE  = 3
	NORMALSTOP_STATE       = 4
	CHECKPEOPEN_STATE      = 5
	PEP2OPEN_STATE         = 6
	WARNB4OPEN_STATE       = 7
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
	DELAYPE_STATE          = 19
	REVERSE2CLOSE_STATE    = 20
	REVERSE2OPEN_STATE     = 21
	SAFETYSTOP_STATE       = 22
	ENTRAPMENTSTOP_STATE   = 23
	FAULT1_STATE           = 24
	FAULT2_STATE           = 25
	FAULT3_STATE           = 26
	FAULT4_STATE           = 27
	FAULT5_STATE           = 28
	FAULT7_STATE           = 29
	FAULT8_STATE           = 30
	FAULT14_STATE          = 31
	FAULT15_STATE          = 32
	ERROR1_STATE           = 33
	ERROR2_STATE           = 34
	ERROR6_STATE           = 35
	ERROR8_STATE           = 36
	ERROR9_STATE           = 37
	ERROR10_STATE          = 38
	ERROR12_STATE          = 39
	ERROR13_STATE          = 40
	ALERT1_STATE           = 41
	ALERT2_STATE           = 42
	ALERT4_STATE           = 43
	ALERT5_STATE           = 44
	ALERT6_STATE           = 45
	ALERT21_STATE          = 46
	FACTORY_TEST_STATE     = 47
	LEARNLIMIT_GEO_STATE   = 48
	LEARNLIMIT_GEC_STATE   = 49
)

// battery states (3.2.2.3)
const (
	BATTERY_DEAD            = 0
	BATTERY_DEAD_OPEN_GATE  = 1
	BATTERY_CONSERVE_LEVEL2 = 2
	BATTERY_CONSERVE_LEVEL1 = 3
	BATTERY_OK              = 4
)
