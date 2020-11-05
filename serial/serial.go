package serial

import (
	//	gser "github.com/goburrow/serial"
	"fmt"
	"github.com/tarm/serial"
)

type serialLayer struct {
	//port gser.Port
	serial *serial.Port
}

func New(addr string) (l *serialLayer, err error) {
	l = &serialLayer{}

	fmt.Printf("opening serial: %s\n", addr)

	s, err := serial.OpenPort(&serial.Config{Name: addr, Baud: 38400})
	if err != nil {
		return nil, err
	}

	l.serial = s

	// port implements io.ReadWriteCloser
	// port, err := gser.Open(&gser.Config{
	// 	Address:  addr,
	// 	BaudRate: 38400,
	// 	DataBits: 8,
	// 	Parity:   "N",
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// l = &serialLayer{
	// 	port: port,
	// }
	return l, nil
}

func (l *serialLayer) Write(b []byte) (int, error) {
	return l.serial.Write(b)
}

func (l *serialLayer) Read(b []byte) (int, error) {
	return l.serial.Read(b)
}

func (l *serialLayer) Close() error {
	return l.serial.Close()
}
