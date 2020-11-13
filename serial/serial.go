package serial

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/tarm/serial"
	"io"
	"time"
)

type serialLayer struct {
	addr    string
	verbose bool
	port    io.ReadWriteCloser
	reader  io.ByteReader
}

func New(addr string, verbose bool) (l *serialLayer, err error) {
	l = &serialLayer{
		addr:    addr,
		verbose: verbose,
	}

	return l, nil
}

func (l *serialLayer) Open() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	errChan := make(chan error)

	go func() {
		l.debug("opening serial: %s", l.addr)

		s, err := serial.OpenPort(&serial.Config{
			Name: l.addr,
			Baud: 38400,
		})
		if err != nil {
			errChan <- err
			return
		}

		l.port = s
		l.reader = bufio.NewReader(s)

		l.debug("serial ready")
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *serialLayer) Write(b []byte) (int, error) {
	return l.port.Write(b)
}

func (l *serialLayer) Read(b []byte) (int, error) {
	return l.port.Read(b)
}

func (l *serialLayer) Close() error {
	defer l.debug("serial closed")
	return l.port.Close()
}

func (l *serialLayer) ReadByte() (byte, error) {
	return l.reader.ReadByte()
}

func (l *serialLayer) debug(msg string, args ...interface{}) {
	if l.verbose {
		color.Blue(fmt.Sprintf("[S] "+msg+"\n", args...))
	}
}
