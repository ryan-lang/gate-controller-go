package logical

import (
	"fmt"
)

type Packet struct {
	SOM         byte
	Address     byte
	MessageSize byte
	MessageType byte
	Message     []byte
	Checksum    byte
}

func NewPacket(msgType byte, msg []byte, addr byte) *Packet {
	p := &Packet{
		SOM:         0xFF,
		Address:     addr,
		MessageSize: byte(len(msg) + 1),
		MessageType: msgType,
		Message:     msg,
	}

	p.Checksum = calcChecksum(p.Frame())
	return p
}

func (p *Packet) Bytes() []byte {
	return p.Frame()
}

func (p *Packet) Header() []byte {
	return []byte{p.SOM, p.Address, p.MessageSize}
}

func (p *Packet) MessageBlock() []byte {
	m := []byte{}
	m = append(m, p.MessageType)
	m = append(m, p.Message...)
	return m
}

func (p *Packet) Frame() []byte {
	header := p.Header()
	msg := p.MessageBlock()

	f := []byte{}
	f = append(f, header...)
	f = append(f, msg...)
	f = append(f, p.Checksum)

	return f
}

func (p *Packet) ValidateChecksum() error {
	var sum int
	for _, b := range p.Frame() {
		sum += int(b)
	}

	if sum%256 == 0 {
		return nil
	} else {
		return &ErrInvalidChecksum{}
	}
}

func (p *Packet) Validate() error {

	// address must be <=254
	if int(p.Address) > 254 {
		return &ErrInvalidAddressByte{p.Address}
	}

	// size must be <=254
	if int(p.MessageSize) > 254 {
		return &ErrInvalidMessageSizeByte{p.MessageSize}
	}

	// ascii chars <=127
	if int(p.MessageType) > 127 {
		return &ErrInvalidMessageTypeByte{p.MessageType}
	}

	for _, m := range p.Message {
		if int(m) > 127 {
			return &ErrInvalidMessageByte{m}
		}
	}

	return nil
}

func calcChecksum(frame []byte) byte {
	var sum int
	for _, f := range frame {
		sum += int(f)
		fmt.Printf("frame %x sum: %d\n", f, sum)
	}
	mod := sum % 256
	fmt.Printf("mod: %d\n", mod)
	comp := ^mod
	fmt.Printf("comp: %d\n", comp)
	chksum := comp + 1
	fmt.Printf("checksum: %d\n", chksum)

	return byte(chksum)
}
