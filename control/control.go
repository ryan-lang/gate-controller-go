package control

import (
	"context"
	"fmt"
	msgs "gate/control/messages"
	"gate/logical"
	"github.com/fatih/color"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// serial -> logical -> protocol -> control

type controlLayer struct {
	verbose bool
	logical LogicalLayer
	txPend  Transaction
	txCtr   int

	txLock  sync.RWMutex
	msgLock sync.Mutex

	closeChan chan bool
	writeBuf  chan *logical.Packet
}

type (
	LogicalLayer interface {
		Start(chan<- *logical.Packet, <-chan *logical.Packet, chan<- error) error
		Close()
	}
)

func New(
	l LogicalLayer,
	verbose bool,
) *controlLayer {
	c := &controlLayer{
		verbose:   verbose,
		logical:   l,
		txPend:    nil,
		txCtr:     0,
		closeChan: make(chan bool),
		writeBuf:  make(chan *logical.Packet),
	}

	return c
}

// start reading from logical ReadChan
// will block until cancel is called
func (c *controlLayer) Run() error {
	c.debugn("starting controller")
	defer c.debugn("controller closed")

	packetRead := make(chan *logical.Packet)
	packetWrite := make(chan *logical.Packet)
	packetErr := make(chan error)

	defer close(packetWrite)

	var g run.Group

	g.Add(func() error {
		return c.logical.Start(packetRead, packetWrite, packetErr)
	}, func(err error) {
		c.logical.Close()
	})

	g.Add(func() error {
		for {
			select {
			case <-c.closeChan:
				return nil

			// read a packet from logical
			case p, ok := <-packetRead:
				if !ok {
					return nil
				}
				c.handlePacketRead(p)
				continue

			// handle a packet error (read or write)
			case err, ok := <-packetErr:
				if !ok {
					return nil
				}
				c.handlePacketError(err)
				continue

			case p, ok := <-c.writeBuf:
				if !ok {
					return nil
				}
				c.debugn("writing packet")
				select {
				case packetWrite <- p:
					c.debugn("wrote packet")
					continue
				case <-c.closeChan:
					return nil
				}
			}
		}
	}, func(error) {
		close(packetRead)
		close(packetErr)
	})

	return g.Run()
}

func (c *controlLayer) Close() {
	close(c.closeChan)
}

// sends reset to gate, no ack
func (c *controlLayer) Reset(ctx context.Context, addr int) error {
	msg := msgs.NewResetMessage(addr)

	return c.resolveMsg(ctx, msg, time.Second*5)
}

func (c *controlLayer) Version(ctx context.Context, addr int) (*msgs.VersionResponse, error) {
	msg := msgs.NewVersionMessage(addr)

	res, err := c.resolveTx(ctx, msg, time.Second*5)
	if err != nil {
		return nil, err
	}
	return res.(*msgs.VersionResponse), err
}

func (c *controlLayer) GateControl(ctx context.Context, req *msgs.GateControlRequest) error {
	msg := msgs.NewGateControlMessage(req)

	_, err := c.resolveTx(ctx, msg, time.Second*5)
	return err
}

func (c *controlLayer) GateStatus(ctx context.Context, addr int) (*msgs.GateStatusResponse, error) {
	msg := msgs.NewGateStatusMessage(addr)

	res, err := c.resolveTx(ctx, msg, time.Second*5)
	if err != nil {
		return nil, err
	}
	return res.(*msgs.GateStatusResponse), err
}

func (c *controlLayer) GateFault(ctx context.Context, addr int) (*msgs.GateFaultResponse, error) {
	msg := msgs.NewGateFaultMessage(addr)

	res, err := c.resolveTx(ctx, msg, time.Second*5)
	if err != nil {
		return nil, err
	}

	r := res.(*msgs.GateFaultResponse)
	return r, err
}

func (c *controlLayer) handlePacketRead(p *logical.Packet) {
	c.debugn("[read] addr=%d, type=%s, msg=% x", p.Address, string(p.MessageType), p.Message)

	// return ErrUnsolicitedFrame if we never made a request
	if c.txPend == nil {
		c.warn("unsolicited frame...")
		return
	}

	// return ErrUnexpectedFrame if we made a request, and this is the wrong reply
	if !c.txPend.FilterResponse(p) {
		c.txPend.ErrChan() <- &ErrUnexpectedFrame{}
	}

	// completes the transaction, sends on tx ResponseChan
	c.txPend.HandleResponse(p)
	return
}

func (c *controlLayer) handlePacketError(err error) {
	if c.txPend == nil {
		c.warn("packet error: %s", err.Error())
		return
	}

	c.txPend.ErrChan() <- err
	return
}

func (c *controlLayer) txRecvTimeout(ctx context.Context, t Transaction, timeout time.Duration) (interface{}, error) {
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	//recv response, recv error, timeout, or cancel
	select {
	case res := <-t.ResponseChan():
		return res, nil
	case err := <-t.ErrChan():
		return nil, err
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "tx recv cancelled")
	case <-rctx.Done():
		return nil, errors.Wrap(rctx.Err(), "tx recv timed out")
	}
}

func (c *controlLayer) resolveMsg(ctx context.Context, m Message, timeout time.Duration) error {
	c.debugn("[msg] resolving with timeout %s", timeout)

	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.msgLock.Lock()
	defer c.msgLock.Unlock()
	c.debugn("[msg] got lock")

	// write, timeout, or cancel
	select {
	case c.writeBuf <- m.Packet():
		return nil
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "tx write cancelled")
	case <-rctx.Done():
		return errors.Wrap(rctx.Err(), "tx write timed out")
	}
}

func (c *controlLayer) resolveTx(ctx context.Context, t Transaction, timeout time.Duration) (res interface{}, err error) {
	c.debugn("[tx%d] resolving with timeout %s", c.txCtr, timeout)

	mainTimer := newTimer()
	var waitWriteTime, waitRespTime time.Duration
	defer func() {
		c.debugn("[tx%d] done: waitWrite=%s, waitResp=%s, total=%s",
			c.txCtr, waitWriteTime, waitRespTime, mainTimer.Since())
		if err != nil {
			c.error("[tx%d] tx failed: %s", c.txCtr, err.Error())
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.txLock.Lock()
	defer c.txLock.Unlock()

	// store pending transaction
	c.txPend = t
	c.txCtr = c.txCtr + 1
	if c.txCtr >= 255 {
		c.txCtr = 1
	}

	// up to 3 retries
	timer := newTimer()
	for i := 0; i < 3; i++ {
		if i > 0 {
			c.warn("[tx%d] retry %d", c.txCtr, i)
		}

		// write the message
		c.debugn("[tx%d] wait write", c.txCtr)
		err = c.resolveMsg(ctx, t, time.Millisecond*500)
		if err != nil {
			c.warn("[tx%d] write failed: %s", c.txCtr, err.Error())
			continue
		}
		waitWriteTime = timer.Since()

		// recv response values
		c.debugn("[tx%d] wait recv", c.txCtr)
		res, err = c.txRecvTimeout(ctx, t, time.Millisecond*500)
		if err == nil {
			break
		}
		c.warn("[tx%d] recv failed: %s", c.txCtr, err.Error())
	}
	waitRespTime = timer.Since()

	return
}

func (c *controlLayer) info(msg string, args ...interface{}) {
	if c.verbose {
		color.Green(fmt.Sprintf("[C] "+msg+"\n", args...))
	}
}

func (c *controlLayer) debug(msg string, args ...interface{}) {
	if c.verbose {
		color.Cyan(fmt.Sprintf("[C] "+msg, args...))
	}
}

func (c *controlLayer) debugn(msg string, args ...interface{}) {
	if c.verbose {
		color.Cyan(fmt.Sprintf("[C] "+msg+"\n", args...))
	}
}

func (c *controlLayer) warn(msg string, args ...interface{}) {
	if c.verbose {
		color.Magenta(fmt.Sprintf("[C] "+msg+"\n", args...))
	}
}

func (c *controlLayer) error(msg string, args ...interface{}) {
	if c.verbose {
		color.Red(fmt.Sprintf("[C] "+msg+"\n", args...))
	}
}

func (c *controlLayer) debugTimer(t time.Time, n string) {
	c.info("%s took %s", n, time.Since(t))
}

type timeTracker struct {
	time time.Time
}

func newTimer() *timeTracker {
	return &timeTracker{
		time: time.Now(),
	}
}

func (t *timeTracker) Since() time.Duration {
	return time.Since(t.time)
}
