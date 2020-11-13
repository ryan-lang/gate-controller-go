package messages

import (
	"encoding/json"
	"gate/logical"
	"strconv"
)

type GateFaultMessage struct {
	responseChan chan interface{}
	errChan      chan error
	address      int
}

type GateFaultResponse struct {
	NumberOfFaults int16
	FaultCodes     []int16
}

func NewGateFaultMessage(addr int) *GateFaultMessage {
	return &GateFaultMessage{
		responseChan: make(chan interface{}, 1),
		errChan:      make(chan error, 3),
		address:      addr,
	}
}

func (m *GateFaultMessage) Packet() *logical.Packet {
	return logical.NewPacket(0x46, nil, byte(m.address))
}

func (m *GateFaultMessage) FilterResponse(p *logical.Packet) bool {
	return p.MessageType == 0x46
}

func (m *GateFaultMessage) HandleResponse(p *logical.Packet) error {
	if len(p.Message) == 0 {
		return &ErrInvalidResponse{}
	}

	r := &GateFaultResponse{}

	{
		c, _ := strconv.Atoi(string(p.Message[0:2]))
		r.NumberOfFaults = int16(c)
	}

	pairs := int(len(p.Message[2:]) / 2)

	for i := 0; i < pairs; i++ {
		start := 2 * (i + 1)
		i, _ := strconv.ParseInt(string(p.Message[start:start+2]), 16, 16)
		r.FaultCodes = append(r.FaultCodes, int16(i))
	}

	m.responseChan <- r
	return nil
}

func (m *GateFaultMessage) ResponseChan() chan interface{} {
	return m.responseChan
}

func (m *GateFaultMessage) ErrChan() chan error {
	return m.errChan
}

func (m *GateFaultResponse) Diff(s *GateFaultResponse) bool {
	return m.NumberOfFaults != s.NumberOfFaults ||
		!equal(m.FaultCodes, s.FaultCodes)
}

func (m *GateFaultResponse) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func equal(a, b []int16) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// fault codes (3.2.3.1)

const (
	FAL1       = 0
	FAL2       = 1
	FAL3       = 2
	FAL4       = 3
	ERR1       = 4
	ERR2       = 5
	ERR4       = 7
	ERR6       = 9
	ALE1       = 13
	ALE2       = 14
	ALE3       = 15
	ALE4       = 16
	ALE5       = 17
	ALE6       = 18
	ALE7_ELD   = 19
	ALE8_ELD   = 20
	ALE9_ELD   = 21
	AL10_ELD   = 22
	AL11_ELD   = 23
	AL12_ELD   = 24
	ALE7_IOLD  = 25
	ALE8_IOLD  = 26
	ALE9_IOLD  = 27
	AL10_IOLD  = 28
	AL11_IOLD  = 29
	AL12_IOLD  = 30
	ALE7_OOLD  = 31
	ALE8_OOLD  = 32
	ALE9_OOLD  = 33
	AL10_OOLD  = 34
	AL11_OOLD  = 35
	AL12_OOLD  = 36
	ALE7_SLD   = 37
	ALE8_SLD   = 38
	ALE9_SLD   = 39
	AL10_SLD   = 40
	AL11_SLD   = 41
	AL12_SLD   = 42
	AL13       = 43
	AL14       = 44
	AL15       = 45
	ERR3_ELD   = 58
	ERR3_IOLD  = 59
	ERR3_OOLD  = 60
	ERR3_SLD   = 61
	AL17       = 66
	ERR8       = 81
	ERR9       = 82
	AL18       = 83
	AL19       = 143
	FAL5_OPEN  = 144
	FAL5_CLOSE = 145
	ER10_OPEN  = 146
	ER10_CLOSE = 147
	AL20       = 155
	FAL14      = 166
	AL21       = 167
	ERR11      = 168
	AL22       = 170
	FAL7       = 171
	AL24       = 172
	ER12       = 177
	ER13       = 180
	FAL6       = 181
	FAL8       = 185
	FAL2_S1    = 194
	FAL2_S2    = 195
	FAL2_S3    = 196
	FAL2_S1_2  = 202
	FAL2_S2_2  = 203
	FAL2_S3_2  = 204
	ER14_ELD   = 205
	ER14_IOLD  = 206
	ER14_OOLD  = 207
	ER14_CLD   = 208
	AL26       = 213
	AL27       = 215
)
