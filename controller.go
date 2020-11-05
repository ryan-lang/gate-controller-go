package controller

import (
	"io"
)

type (
	DeviceLogicalLayer interface {
		Read()
		Write()
	}
	DeviceControlLayer interface {
		Send()
		Recv()
	}
	ClientControlLayer interface {
		DeviceUpdate()
		RequestUpdate()
	}
	ClientSerialLayer interface {
		Write()
		Read()
	}
)

type Controller interface {
	Reader() io.Reader
	Writer() io.Writer
}

type gateController struct{}
