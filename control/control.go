package control

import (
	"context"
	"fmt"
	msgs "gate/control/messages"
	"gate/logical"
	"sync"
	"time"
)

// serial -> logical -> protocol -> control

type controlLayer struct {
	verbose bool
	logical LogicalLayer
	txPend  Transaction
	lock    sync.Mutex
}

type (
	LogicalLayer interface {
		ReadChan() chan *logical.Packet
		WriteChan() chan *logical.Packet
		ErrChan() chan error
		Close() error
	}
)

func New(
	logical LogicalLayer,
) *controlLayer {
	return &controlLayer{
		verbose: true,
		logical: logical,
		txPend:  nil,
	}
}

// start reading from logical ReadChan
func (c *controlLayer) Start() error {
	for {
		select {
		case p := <-c.logical.ReadChan():
			err := c.handleRead(p)
			if err != nil {
				c.warnn(err.Error())
			}
		case err := <-c.logical.ErrChan():
			if err != nil {
				c.warnn(err.Error())
			}
		}
	}

	return nil
}

// sends reset to gate, no ack
func (c *controlLayer) Reset(ctx context.Context) error {
	msg := msgs.NewResetMessage()

	err := c.writeMessage(msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *controlLayer) Version(ctx context.Context) (*msgs.VersionResponse, error) {
	msg := msgs.NewVersionMessage()

	res, err := c.writeTransactionTimeout(ctx, msg)
	if err != nil {
		return nil, err
	}

	vres := res.(*msgs.VersionResponse)
	return vres, nil
}

func (c *controlLayer) GateControl(ctx context.Context, req *msgs.GateControlRequest) error {
	msg := msgs.NewGateControlMessage(req)

	_, err := c.writeTransactionTimeout(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *controlLayer) GateStatus(ctx context.Context) (*msgs.GateStatusResponse, error) {
	msg := msgs.NewGateStatusMessage()

	res, err := c.writeTransactionTimeout(ctx, msg)
	if err != nil {
		return nil, err
	}

	vres := res.(*msgs.GateStatusResponse)
	return vres, nil
}

func (c *controlLayer) GateFault(ctx context.Context) (*msgs.GateFaultResponse, error) {
	msg := msgs.NewGateFaultMessage()

	res, err := c.writeTransactionTimeout(ctx, msg)
	if err != nil {
		return nil, err
	}

	vres := res.(*msgs.GateFaultResponse)
	return vres, nil
}

func (c *controlLayer) writeMessage(m Message) error {
	c.debugn("[wait] control lock")
	c.lock.Lock()
	defer c.lock.Unlock()

	c.write(m.Packet())
	return nil
}

func (c *controlLayer) writeTransactionTimeout(ctx context.Context, t Transaction) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	d := make(chan interface{}, 1)
	e := make(chan error, 1)

	go func() {
		res, err := c.writeTransaction(t)
		if err != nil {
			e <- err
		}

		d <- res
	}()

	select {
	case res := <-d:
		return res, nil
	case err := <-e:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *controlLayer) writeTransaction(t Transaction) (interface{}, error) {
	c.debugn("[wait] control lock")
	c.lock.Lock()
	defer c.lock.Unlock()

	c.txPend = t

	// write the packet
	c.debugn("[wait] logical write")
	c.write(t.Packet())

	// wait for transactional response
	c.debugn("[wait] tx resp")
	res := <-t.ResponseChan()

	return res, nil
}

func (c *controlLayer) write(p *logical.Packet) {
	c.debugn("[write] addr=%d, type=%d, checksum=%x, msg=% x", p.Address, p.MessageType, p.Checksum, p.Message)
	c.logical.WriteChan() <- p
}

func (c *controlLayer) handleRead(p *logical.Packet) error {
	c.debugn("[read] addr=%d, type=%d, msg=% x", p.Address, p.MessageType, p.Message)

	// return ErrUnsolicitedFrame if we never made a request
	if c.txPend == nil {
		return &ErrUnsolicitedFrame{}
	}

	// return ErrUnexpectedFrame if we made a request, and this is the wrong reply
	if !c.txPend.FilterResponse(p) {
		return &ErrUnexpectedFrame{}
	}

	// completes the transaction
	err := c.txPend.HandleResponse(p)
	if err != nil {
		return err
	}

	return nil
}

func (c *controlLayer) debug(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}

func (c *controlLayer) debugn(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func (c *controlLayer) warnn(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}
