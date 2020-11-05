package control_test

import (
	"context"
	"gate/control"
	"gate/control/messages"
	"gate/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	//	"fmt"
)

func TestReset(t *testing.T) {
	done := make(chan bool)

	l := NewMockLogical()
	c := control.New(l)
	go c.Start()

	// logical listener
	go func() {
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x52))
		done <- true
	}()

	err := c.Reset(context.Background())
	require.Nil(t, err)

	<-done
}

func TestVersion(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l)
	go c.Start()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x56))

		// simulate response from logical
		l.readChan <- &logical.Packet{
			MessageType: 0x56,
			Message:     []byte{0x68, 0x34, 0x2e, 0x31, 0x38, 0x2c, 0x20, 0x42, 0x6f, 0x6f, 0x74, 0x20, 0x4c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x3a, 0x20, 0x56, 0x31, 0x2e, 0x30},
		}
	}()

	res, err := c.Version(context.Background())
	require.Nil(t, err)
	require.Equal(t, "h4.18, Boot Loader: V1.0", res.Version)
}

func TestGateControl(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l)
	go c.Start()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x43))

		// simulate response from logical
		l.readChan <- &logical.Packet{
			MessageType: 0x43,
		}
	}()

	err := c.GateControl(context.Background(), &messages.GateControlRequest{
		PushButtonOpen: true,
	})
	require.Nil(t, err)
}

func TestGateStatus(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l)
	go c.Start()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x53))

		// simulate response from logical
		l.readChan <- &logical.Packet{
			MessageType: 0x53,
			Message:     []byte{0x001b},
		}
	}()

	res, err := c.GateStatus(context.Background())
	require.Nil(t, err)
	require.Equal(t, uint16(27), res.LastCommandStatus)
}

func TestGateFault(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l)
	go c.Start()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x46))

		// simulate response from logical
		l.readChan <- &logical.Packet{
			MessageType: 0x46,
			Message:     []byte{4},
		}
	}()

	res, err := c.GateFault(context.Background())
	require.Nil(t, err)
	require.Equal(t, uint8(4), res.NumberOfFaults)
}

type mockLogical struct {
	readPacket *logical.Packet
	readChan   chan *logical.Packet
	writeChan  chan *logical.Packet
}

func NewMockLogical() *mockLogical {
	l := &mockLogical{
		readChan:  make(chan *logical.Packet),
		writeChan: make(chan *logical.Packet),
	}

	return l
}

func (s *mockLogical) Read() (packet *logical.Packet, err error) {
	return s.readPacket, nil
}

func (s *mockLogical) Write([]byte) (int, error) {
	return 0, nil
}

func (s *mockLogical) Close() error {
	return nil
}

func (s *mockLogical) ReadChan() chan *logical.Packet {
	return s.readChan
}

func (s *mockLogical) WriteChan() chan *logical.Packet {
	return s.writeChan
}
