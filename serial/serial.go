package serial

import (
	"bufio"
	"gousb"
)

type serialLayer struct {
	bytes.Buffer
	device *gousb.Device
	ctx    context.Context
}

func New() (l *serialLayer, err error) {
	l = &serialLayer{}

	l.ctx = gousb.NewContext()

	l.dev, err = ctx.OpenDeviceWithVIDPID(0x046d, 0xc526)
	if err != nil {
		return nil, err
	}

	l.intf, l.done, err = dev.DefaultInterface()
	if err != nil {
		return nil, err
	}
}

func (l *serialLayer) Write(b []byte) (int, error) {
	numBytes, err := ep.Write(data)
	if numBytes != 5 {
		log.Fatalf("%s.Write([5]): only %d bytes written, returned error is %v", ep, numBytes, err)
	}
}

func (l *serialLayer) Read(b []byte) (int, error) {

}

func (l *serialLayer) Close() error {
	l.dev.Close()
	l.context.Close()
}
