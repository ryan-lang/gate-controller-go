package logical

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	//	"github.com/howeyc/crc16"
	"io"
)

type logicalLayer struct {
	verbose   bool
	serial    SerialLayer
	readChan  chan *Packet
	writeChan chan *Packet
	errChan   chan error
}

type (
	SerialLayer interface {
		io.ReadWriteCloser
	}
)

func New(
	serial SerialLayer,
) *logicalLayer {
	return &logicalLayer{
		verbose:   true,
		serial:    serial,
		readChan:  make(chan *Packet),
		writeChan: make(chan *Packet),
		errChan:   make(chan error),
	}
}

// converts our io stream into reader/writer channels
// and manages reading/writing in two gofuncs
func (l *logicalLayer) Start() {
	go l.reader()
	go l.writer()
}

func (l *logicalLayer) ReadChan() chan *Packet {
	return l.readChan
}

func (l *logicalLayer) WriteChan() chan *Packet {
	return l.writeChan
}

func (l *logicalLayer) ErrChan() chan error {
	return l.errChan
}

func (l *logicalLayer) Close() error {
	close(l.writeChan)
	close(l.readChan)
	close(l.errChan)

	return l.serial.Close()
}

func (l *logicalLayer) reader() {
	reader := bufio.NewReader(l.serial)

	for {
		p, err := l.read(reader)
		if err != nil {
			l.Debug("err")
			l.errChan <- err
		}

		l.readChan <- p
	}
}

func (l *logicalLayer) writer() {
	writer := bufio.NewWriter(l.serial)

	for {
		select {
		case p := <-l.writeChan:
			err := l.write(writer, p)
			if err != nil {
				l.Debug("err")
				l.errChan <- err
			}
		}
	}
}

func (l *logicalLayer) write(writer *bufio.Writer, p *Packet) error {
	b := p.Bytes()
	l.Debug("[write] % x", b)

	_, err := writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// read the next full packet in the serial stream; may return packet
// even if err != nil
func (l *logicalLayer) read(reader *bufio.Reader) (packet *Packet, err error) {

	packet = &Packet{}

	// read the first byte
	for {
		l.Debug("reading next byte, size=%d, buf=%d, %x", reader.Size(), reader.Buffered())
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		l.Debug("read byte: %x", b)

		// first byte must be 0xFF or we ignore
		if b == 0xFF {
			packet.SOM = b
			l.Debug("0:SOM=%x", packet.SOM)

			break
		} else {
			l.Debug("0?:not SOM=%x", b)
		}
	}

	// read remaining header (SOM + 2 bytes)
	{
		err = l.peekValidateNext(reader, 2)
		if err != nil {
			return packet, err
		}

		head := make([]byte, 2)
		_, err = io.ReadFull(reader, head)
		if err != nil {
			return packet, err
		}

		packet.Address = head[0]
		packet.MessageSize = head[1]

		l.Debug("1:addr=%x", packet.Address)
		l.Debug("2:msgSize=%d", packet.MessageSize)

		// ++ VALIDATE
		err = packet.Validate()
		if err != nil {
			return packet, err
		}
	}

	// read next byte = messageId
	{
		err = l.peekValidateNext(reader, 1)
		if err != nil {
			return packet, err
		}

		msgID, err := reader.ReadByte()
		if err != nil {
			return packet, err
		}

		packet.MessageType = msgID
		l.Debug("3:msgType=%x", packet.MessageType)

		// ++ VALIDATE
		err = packet.Validate()
		if err != nil {
			return packet, err
		}
	}

	// read next size-1 bytes (size includes msgID in prev byte)
	{
		err = l.peekValidateNext(reader, int(packet.MessageSize)-1)
		if err != nil {
			return packet, err
		}

		msg := make([]byte, packet.MessageSize-1)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			return packet, err
		}

		packet.Message = msg
		l.Debug("4-%d:msg=% x", len(packet.Message)+3, packet.Message)
	}

	// read & validate checksum
	{
		err = l.peekValidateNext(reader, 1)
		if err != nil {
			return packet, err
		}

		checksum, err := reader.ReadByte()
		if err != nil {
			return packet, err
		}
		packet.Checksum = checksum
		l.Debug("%d:checksum=%x", len(packet.Message)+4, checksum)

		err = packet.ValidateChecksum()
		if err != nil {
			l.Debug("frame=% x checksum=bad\n", packet.Frame())
			return packet, err
		}
	}

	l.Debug("frame=% x checksum=ok\n", packet.Frame())
	return packet, nil
}

// read the next n bytes from the stream. rewind and return error
// if we saw a bad byte
func (l *logicalLayer) peekValidateNext(reader *bufio.Reader, n int) error {
	bs, err := reader.Peek(n)
	if err != nil {
		return err
	}
	for _, b := range bs {
		if b == 0xFF {
			return errors.Wrap(&ErrUnexpectedByte{b}, fmt.Sprintf("peeked invalid byte in next %d bytes", n))
		}
	}
	return nil
}

func (l *logicalLayer) Debug(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}
