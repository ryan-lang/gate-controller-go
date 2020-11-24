package logical

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"io"
	"os"
	"sync"
	"time"
)

type logicalLayer struct {
	verbose   bool
	serial    SerialLayer
	closeChan chan bool
}

type (
	SerialLayer interface {
		Open() error
		io.ReadWriteCloser
		io.ByteReader
	}
)

func New(
	serial SerialLayer,
	verbose bool,
) *logicalLayer {
	return &logicalLayer{
		verbose: verbose,
		serial:  serial,
	}
}

// opens serial connection, starts reader & writer gofuncs
// blocks until Stop is called
func (l *logicalLayer) Start(
	readChan chan<- *Packet,
	writeChan <-chan *Packet,
	errChan chan<- error,
) error {
	l.Debug("starting logical")

	l.closeChan = make(chan bool)

	var wg sync.WaitGroup

	err := l.serial.Open()
	if err != nil {
		return errors.Wrap(err, "failed to open serial port")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		wg.Add(1)
		defer wg.Done()

		l.reader(ctx, readChan, errChan)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()

		l.writer(ctx, writeChan, errChan)
	}()

	select {
	case <-l.closeChan:
		cancel()

		l.Debug("waiting for reader/writer...")
		wg.Wait()

		l.Debug("waiting for serial close...")
		err := l.serial.Close()
		if err != nil {
			return err
		}

		l.Debug("logical layer closed")
		return nil
	}
}

func (l *logicalLayer) Close() {
	close(l.closeChan)
}

func (l *logicalLayer) reader(
	ctx context.Context,
	readChan chan<- *Packet,
	errChan chan<- error,
) {
	serialByte := make(chan byte)
	serialErr := make(chan error)

	// start the first-byte reader, which pushes
	// next byte onto the channel, or pushes error
	go func() {
		for {
			b, err := l.serial.ReadByte()
			if err != nil {
				if err == io.EOF {
					continue
				}

				switch err := err.(type) {
				case *os.PathError:
					if err.Err == os.ErrClosed {
						// ignore "file already closed errors"
						// this error means shutting down
						return
					}
				default:
				}

				serialErr <- err
				continue
			}

			serialByte <- b
		}
	}()

	// main loop looks for SOM bytes, or error or cancel
	for {
		select {

		// exit if main loop is cancelled
		case <-ctx.Done():
			return

		// propagate main loop serial errors
		case err := <-serialErr:
			errChan <- err
			continue

		// read the next byte
		case b := <-serialByte:

			// ingore any bytes that are not SOM
			if b != 0xFF {
				continue
			}

			// inner loop looks for non-SOM bytes, or error, or cancel
			var packet *Packet
		packetRead:
			for {
				if packet == nil {
					l.debugRecv("new packet")
					packet = &Packet{
						SOM: 0xFF,
					}
				}

				// start a stricter timeout for reading the next byte
				bctx, bcancel := context.WithTimeout(ctx, time.Millisecond*500)

				select {

				// main loop cancelled, exit
				case <-ctx.Done():
					return

				// next-byte reader cancelled, return to main loop
				case <-bctx.Done():
					errChan <- errors.Wrap(bctx.Err(), "timed out waiting for next byte")
					break packetRead

				// if serial read error, return to main loop
				case err := <-serialErr:
					errChan <- err
					break packetRead

				// next byte for packet
				case b := <-serialByte:
					// SOM byte mid-packet, reset the packet
					if b == 0xFF {
						errChan <- &ErrUnexpectedByte{}
						packet = nil
						continue
					}

					if packet.Address == 0 {
						packet.Address = b

					} else if packet.MessageSize == 0 {
						packet.MessageSize = b

					} else if packet.MessageType == 0 {
						packet.MessageType = b

					} else if len(packet.Message) < (int(packet.MessageSize) - 1) {
						packet.Message = append(packet.Message, b)

					} else if packet.Checksum == 0 {

						packet.Checksum = b

						err := packet.ValidateChecksum()
						if err != nil {
							l.debugRecv("frame=% x checksum=bad", packet.Frame())

							errChan <- err
							break packetRead
						}

						l.debugRecv("frame=% x checksum=ok", packet.Frame())
					}

					// validate full packet on each iteration
					// give up on packet if validation fails
					err := packet.Validate()
					if err != nil {
						errChan <- err
						break packetRead
					}

					// packet is complete & valid
					if packet.Checksum > 0 {
						l.debugRecv("readchan blocked")
						readChan <- packet
						l.debugRecv("readchan recv")
						break packetRead
					}
				}

				// cancel context
				bcancel()
			}

			l.debugRecv("finished packet")
		}
	}
}

func (l *logicalLayer) writer(
	ctx context.Context,
	writeChan <-chan *Packet,
	errChan chan<- error,
) {
	for {
		select {
		case <-ctx.Done():
			l.debugSend("writer cancelled")
			return
		case p, ok := <-writeChan:
			if !ok {
				return
			}
			err := l.write(l.serial, p)
			if err != nil {
				l.debugSend("err")
				errChan <- err
			}
		}
	}
}

func (l *logicalLayer) handleSerialError(ctx context.Context, errChan chan error, err error) {

	l.debugRecv("serial err recv")
	select {
	case errChan <- err:
	case <-ctx.Done():
	}
}

// blocking for loop that receives on context Done, errChan, and byteChan, sends on errChan and readChan
func (l *logicalLayer) readPacket(
	ctx context.Context,
	b byte,
	packet *Packet,
	readChan chan<- *Packet,
	errChan chan<- error,
) *Packet {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()

	return packet
}

func (l *logicalLayer) write(writer io.Writer, p *Packet) error {
	b := p.Bytes()
	l.debugSend("frame=% x", b)

	_, err := writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (l *logicalLayer) debugSend(msg string, args ...interface{}) {
	l.Debug("[send] "+msg, args...)
}

func (l *logicalLayer) debugRecv(msg string, args ...interface{}) {
	l.Debug("[recv] "+msg, args...)
}

func (l *logicalLayer) Debug(msg string, args ...interface{}) {
	if l.verbose {
		color.Yellow(fmt.Sprintf("[L] "+msg+"\n", args...))
	}
}
