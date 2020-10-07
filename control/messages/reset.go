package messages

type ResetMessage struct {
}

func NewResetMessage() *ResetMessage {
	return &ResetMessage{}
}

func (m *ResetMessage) Packet() *logical.Packet {
	return &logical.Packet{
		MessageType: "R",
	}
}
