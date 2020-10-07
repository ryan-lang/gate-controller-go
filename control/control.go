package control

import (
	"context"
	"gate/logical"
)

// serial -> logical -> protocol -> control

type controlLayer struct {
	verbose bool
	logical LogicalLayer
	txPend  Transaction
}

type (
	LogicalLayer interface {
		ReadChan() chan *logical.Packet
		WriteChan() chan *logical.Packet
		Close() error
	}
)

func New(
	logical LogicalLayer,
) *controlLayer {
	return &controlLayer{
		logical: logical,
		txPend:  nil,
	}
}

func (c *controlLayer) Read() (Message, error) {
	select {
	case p := <-c.logical.ReadChan():

		// return ErrUnsolicitedFrame if we never made a request
		if c.txPend == nil {
			return nil, &ErrUnsolicitedFrame{}
		}

		// return ErrUnexpectedFrame if we made a request, and this is the wrong reply
		if !c.txPend.FilterResponse(p) {
			return nil, &ErrUnexpectedFrame{}
		}

		// completes the transaction
		err = c.txPend.HandleRespone(p)
		if err != nil {
			return err
		}
	}

	return nil
}

// sends reset to gate, no ack
func (c *controlLayer) Reset() error {
	msg := msgs.NewResetMessage()

	err := c.writeMessage(msg)
	if err != nil {
		return err
	}
}

func (c *controlLayer) Version() (*msgs.VersionResponse, error) {
	msg := msgs.NewVersionMessage()

	res, err := c.writeTransaction(msg)
	if err != nil {
		return err
	}

	vres := res.(*msgs.VersionResponse)
	return vres, nil
}

func (c *controlLayer) GateControl(req *msgs.GateControlRequest) error {
	msg := msgs.NewGateControlMessage(req)

	_, err := c.writeTransaction(msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *controlLayer) GateStatus() (*msgs.GateStatusResponse, error) {
	msg := msgs.NewGateStatusMessage()

	res, err := c.writeTransaction(msg)
	if err != nil {
		return err
	}

	vres := res.(*msgs.GateStatusResponse)
	return vres, nil
}

func (c *controlLayer) GateFault() (*msgs.GateFaultResponse, error) {
	msg := msgs.NewGateFaultMessage()

	res, err := c.writeTransaction(msg)
	if err != nil {
		return err
	}

	vres := res.(*msgs.GateFaultResponse)
	return vres, nil
}

func (c *controlLayer) writeMessage(m Message) error {
	err := c.waitReadyWrite()
	if err != nil {
		return err
	}

	c.logical.WriteChan() <- m.Packet()
}

func (c *controlLayer) writeTransaction(t Transaction) (interface{}, error) {
	err := c.waitReadyWrite()
	if err != nil {
		return err
	}

	// TODO: lock
	c.txPend = t

	// write the packet
	c.logical.WriteChan() <- t.Packet()

	res := <-t.ResponseChan()
	return res, nil
}

func (c *controlLayer) isPendingResponse() bool {
	// TODO: read lock
	return c.txPend
}

// TODO: timeout
func (c *controlLayer) waitReadyWrite() {
	for {
		// no writes if waiting on response
		if !c.isPendingResponse() {
			break
		}
	}
}
