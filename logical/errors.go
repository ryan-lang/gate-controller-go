package logical

import (
	"fmt"
)

type (
	ErrInvalidAddressByte struct {
		addr byte
	}
	ErrUnexpectedByte struct {
		byte byte
	}
	ErrInvalidChecksum        struct{}
	ErrInvalidMessageSizeByte struct {
		size byte
	}
	ErrInvalidMessageTypeByte struct {
		typ byte
	}
	ErrInvalidMessageByte struct {
		byte byte
	}
)

func (e *ErrInvalidAddressByte) Error() string {
	return fmt.Sprintf("invalid address=%x", e.addr)
}

func (e *ErrUnexpectedByte) Error() string {
	return fmt.Sprintf("unexpected byte: %x", e.byte)
}

func (e *ErrInvalidChecksum) Error() string {
	return "invalid checksum"
}

func (e *ErrInvalidMessageSizeByte) Error() string {
	return fmt.Sprintf("invalid message size=%d", e.size)
}

func (e *ErrInvalidMessageTypeByte) Error() string {
	return fmt.Sprintf("invalid message type=%d", e.typ)
}

func (e *ErrInvalidMessageByte) Error() string {
	return fmt.Sprintf("invalid message byte=%d", e.byte)
}
