package control_test

import (
	"context"
	"fmt"
	"gate/control"
	"gate/control/messages"
	"gate/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReset(t *testing.T) {
	done := make(chan bool)

	l := NewMockLogical()
	c := control.New(l, true)
	go c.Run()

	// logical listener
	go func() {
		serialWrite := <-l.writeChan
		fmt.Println("got")
		assert.Equal(t, serialWrite.MessageType, byte(0x52))
		done <- true
	}()

	err := c.Reset(context.Background(), 1)
	require.Nil(t, err)

	<-done
}

func TestVersion(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l, true)
	go c.Run()

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

	res, err := c.Version(context.Background(), 1)
	require.Nil(t, err)
	require.Equal(t, "h4.18, Boot Loader: V1.0", res.Version)
}

func TestGateControl(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l, true)
	go c.Run()

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
	c := control.New(l, true)
	go c.Run()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x53))

		p := &logical.Packet{
			MessageType: 0x4e,
			Message: []byte{
				0x30, 0x34, // last cmd status, 0 4
				0x30, 0x34, // curr operator state, 0 4
				0x31,             // faults, 1
				0x34,             // battery state, 4
				0x31,             // ac present, 1
				0x30,             // open limit, 0
				0x31,             // close limit, 1
				0x30,             // partial open limit, 0
				0x30,             // exit loop, 0
				0x30,             // inner obstruction loop, 0
				0x30,             // outer obstruction loop, 0
				0x30,             // reset/shadow loop, 0
				0x30, 0x30, 0x30, // relays, 0 0 0
				0x30,                         // photo eye open, 0
				0x30,                         // photo eye close, 0
				0x30,                         // gate edge both, 0
				0x30,                         // gate edge close, 0
				0x30,                         // gate edge open, 0
				0x30,                         // phto eye both, 0
				0x30, 0x30, 0x30, 0x30, 0x30, // reserved 0 0 0 0 0
				0x30,                               // open too long, 0
				0x30,                               // tailgater, 0
				0x30,                               // loitering, 0
				0x30, 0x30, 0x30, 0x30, 0x31, 0x38, // transient vehicle count, 000018
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, // tenant vehicle count, 000000
				0x30, 0x31, 0x36, 0x32, 0x36, 0x38, // special vehicle count, 016268
				0x30, 0x30, 0x30, 0x33, 0x44, 0x46, // unknown vehicle count, 0003DF
				0x30, 0x32, 0x42, 0x41, 0x31, 0x41, // cycle count, 02BA1A
				0x30, 0x30, 0x30, 0x30, // eld count, 0000
				0x30, 0x30, 0x30, 0x30, // iold count, 0000
				0x30, 0x30, 0x30, 0x30, // oold count, 0000
				0x45, 0x33, 0x32, 0x36, // hld/cld count, E326
				0x30, 0x30, 0x32, 0x38, // transient count, 0028
				0x30, 0x30, 0x30, 0x30, // tenant vends, 0000
				0x35, 0x45, 0x37, 0x35, // special vends, FE75
				0x30, 0x41, 0x36, 0x46, // manual vends, 0A6F
			},
		}

		// simulate response from logical
		l.readChan <- p
	}()

	res, err := c.GateStatus(context.Background(), 1)
	require.Nil(t, err)

	fmt.Println(res.String())
	require.Equal(t, int16(4), res.LastCommandStatus)
	require.Equal(t, int16(4), res.CurrentOperatorState)
	require.Equal(t, true, res.FaultsPresent)
	require.Equal(t, int16(4), res.BatteryState)
	require.Equal(t, true, res.ACPresent)
	require.Equal(t, false, res.OpenLimit)
	require.Equal(t, true, res.CloseLimit)
	require.Equal(t, false, res.PartialOpenLimit)
	require.Equal(t, false, res.ExitLoop)
	require.Equal(t, false, res.InnerObstructionLoop)
	require.Equal(t, false, res.OuterObstructionLoop)
	require.Equal(t, false, res.ResetShadowLoop)
	require.Equal(t, byte(0x30), res.Relay1)
	require.Equal(t, byte(0x30), res.Relay2)
	require.Equal(t, byte(0x30), res.Relay3)
	require.Equal(t, false, res.PhotoEyeOpen)
	require.Equal(t, false, res.PhotoEyeClose)
	require.Equal(t, false, res.GateEdgeBoth)
	require.Equal(t, false, res.GateEdgeClose)
	require.Equal(t, false, res.GateEdgeOpen)
	require.Equal(t, false, res.PhotoEyeBoth)
	require.Equal(t, false, res.OpenTooLong)
	require.Equal(t, false, res.Tailgater)
	require.Equal(t, false, res.Loitering)
}

func TestGateFault(t *testing.T) {
	l := NewMockLogical()
	c := control.New(l, true)
	go c.Run()

	// logical listener
	go func() {
		// wait for write
		serialWrite := <-l.writeChan
		assert.Equal(t, serialWrite.MessageType, byte(0x46))

		p := &logical.Packet{
			MessageType: 0x46,
			Message:     []byte{0x30, 0x31, 0x41, 0x31, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x2d},
		}

		// simulate response from logical
		l.readChan <- p
	}()

	res, err := c.GateFault(context.Background(), 1)
	require.Nil(t, err)

	fmt.Println(res.String())
	require.Equal(t, int16(1), res.NumberOfFaults)
}

type mockLogical struct {
	readPacket *logical.Packet
	readChan   chan *logical.Packet
	writeChan  chan *logical.Packet
	errChan    chan error
}

func NewMockLogical() *mockLogical {
	l := &mockLogical{
		writeChan: make(chan *logical.Packet),
		readChan:  make(chan *logical.Packet),
		errChan:   make(chan error),
	}
	return l
}

func (s *mockLogical) Read() (packet *logical.Packet, err error) {
	return s.readPacket, nil
}

func (s *mockLogical) Write([]byte) (int, error) {
	return 0, nil
}

func (s *mockLogical) Start(
	r chan<- *logical.Packet,
	w <-chan *logical.Packet,
	e chan<- error,
) error {
	for {
		select {
		case p := <-w:
			s.writeChan <- p
		case p := <-s.readChan:
			r <- p
		case err := <-s.errChan:
			e <- err
		}
	}
}

func (s *mockLogical) Close() {
	return
}
