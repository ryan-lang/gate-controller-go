package control

import (
	"gate/logical"
)

type Message interface {
	Packet() *logical.Packet
}

type Transaction interface {
	Message
	// returns true if packet is a matching response to the one
	// we expect
	FilterResponse(*logical.Packet) bool
	HandleResponse(*logical.Packet) error
	ResponseChan() chan interface{}
}
