package service

import (
	"context"
	"fmt"
	msgs "gate/control/messages"
	ops "gate/service/ops"
	"github.com/fatih/color"
	"github.com/oklog/run"
	"github.com/rcrowley/go-metrics"
	"sync"
	"time"
)

type serviceLayer struct {
	verbose    bool
	controller ControlLayer
	addr       int
	done       chan bool

	statuses chan *msgs.GateStatusResponse
	faults   chan *msgs.GateFaultResponse
	errors   chan error

	lock          sync.RWMutex
	listenerCount int
	listeners     map[int]*listener

	metrics *ServiceMetrics
}

type ServiceMetrics struct {
	TxTimer       metrics.Timer
	FaultMeter    metrics.Meter
	ErrMeter      metrics.Meter
	GateUpMeter   metrics.Meter
	GateDownMeter metrics.Meter
}

type ControlLayer interface {
	Run() error
	Close()
	Reset(ctx context.Context, addr int) error
	Version(ctx context.Context, addr int) (*msgs.VersionResponse, error)
	GateControl(ctx context.Context, req *msgs.GateControlRequest) error
	GateStatus(ctx context.Context, addr int) (*msgs.GateStatusResponse, error)
	GateFault(ctx context.Context, addr int) (*msgs.GateFaultResponse, error)
}

type listener struct {
	statuses chan *msgs.GateStatusResponse
	faults   chan *msgs.GateFaultResponse
	errors   chan error
}

func New(
	control ControlLayer,
	verbose bool,
	addr int,
	metrics *ServiceMetrics,
) *serviceLayer {
	return &serviceLayer{
		verbose:    verbose,
		controller: control,
		addr:       addr,
		done:       make(chan bool),
		listeners:  make(map[int]*listener),
		metrics:    metrics,
	}
}

func (s *serviceLayer) Start() error {
	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Add(func() error {
		return s.controller.Run()
	}, func(error) {
		s.controller.Close()
	})

	g.Add(func() error {
		var prevStatus *msgs.GateStatusResponse
		var prevFault *msgs.GateFaultResponse

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.done:
				return nil
			default:

				{
					start := time.Now()
					status, err := s.controller.GateStatus(ctx, s.addr)
					if err == nil {
						s.metrics.TxTimer.UpdateSince(start)
					} else {
						s.metrics.ErrMeter.Mark(1)
						s.publishError(err)
						continue
					}

					if prevStatus == nil || prevStatus.Diff(status) {
						s.publishStatus(status)
					}

					prevStatus = status
				}

				{
					start := time.Now()
					fault, err := s.controller.GateFault(ctx, s.addr)
					if err == nil {
						s.metrics.TxTimer.UpdateSince(start)
					} else {
						s.metrics.ErrMeter.Mark(1)
						s.publishError(err)
						continue
					}

					if prevFault == nil || prevFault.Diff(fault) {
						s.metrics.FaultMeter.Mark(int64(fault.NumberOfFaults))
						s.publishFault(fault)
					}

					prevFault = fault
				}
			}
		}
	}, func(error) {
		cancel()
	})

	return g.Run()
}

func (s *serviceLayer) Close() {
	close(s.done)
}

func (s *serviceLayer) Listen(
	ctx context.Context,
	stCh chan *msgs.GateStatusResponse,
	fCh chan *msgs.GateFaultResponse,
	erCh chan error,
) int {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.listenerCount = s.listenerCount + 1
	listener := &listener{
		statuses: stCh,
		faults:   fCh,
		errors:   erCh,
	}

	s.listeners[s.listenerCount] = listener
	return s.listenerCount
}

func (s *serviceLayer) Unlisten(ctx context.Context, lid int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.listeners[lid]; ok {
		delete(s.listeners, lid)
	}
}

func (s *serviceLayer) PushButtonOpen(ctx context.Context) error {
	defer s.metrics.GateUpMeter.Mark(1)
	return s.Exec(ctx, ops.NewPushButtonOpenOp(s.addr))
}

func (s *serviceLayer) PushButtonClose(ctx context.Context) error {
	defer s.metrics.GateDownMeter.Mark(1)
	return s.Exec(ctx, ops.NewPushButtonCloseOp(s.addr))
}

func (s *serviceLayer) Exec(ctx context.Context, op ops.Operation) error {
	//defer c.debugTimer(time.Now(), op.ID())

	ctx, cancel := context.WithTimeout(ctx, op.Timeout())
	defer cancel()

	creq := op.ControlRequest()

	// make control request
	err := s.controller.GateControl(ctx, creq)
	if err != nil {
		return err
	}

	// start state check loop
	for {
		select {
		case <-ctx.Done():
			s.debug("%s cancelled: "+ctx.Err().Error(), op.ID())
			return ctx.Err()
		default:
			res, err := s.controller.GateStatus(ctx, creq.Address)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", res.String())

			if op.CheckDone(ctx, res) {
				return nil
			}

			fault := op.CheckFault(ctx, res)
			if fault != nil {
				return fault
			}

			if op.CheckInProgress(ctx, res) {
				s.debug("%s in progress - checking again...", op.ID())
			} else {
				s.warn("%s not complete - checking again...", op.ID())
			}
		}
	}
}

func (s *serviceLayer) publishStatus(status *msgs.GateStatusResponse) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var wg sync.WaitGroup

	for _, l := range s.listeners {
		go func(l *listener) {
			wg.Add(1)
			defer wg.Done()
			l.statuses <- status
		}(l)
	}

	wg.Wait()
}

func (s *serviceLayer) publishFault(fault *msgs.GateFaultResponse) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var wg sync.WaitGroup

	for _, l := range s.listeners {
		go func(l *listener) {
			wg.Add(1)
			defer wg.Done()
			l.faults <- fault
		}(l)
	}

	wg.Wait()
}

func (s *serviceLayer) publishError(err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var wg sync.WaitGroup

	for _, l := range s.listeners {
		go func(l *listener) {
			wg.Add(1)
			defer wg.Done()
			l.errors <- err
		}(l)
	}

	wg.Wait()
}

// func (c *serviceLayer) verifyState(
// 	ctx context.Context,
// 	check func(s *msgs.GateStatusResponse) bool,
// 	addr int,
// 	timeout time.Duration,
// 	name string,
// ) error {

// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			c.debugn("%s cancelled: "+ctx.Err().Error(), name)
// 			return ctx.Err()
// 		default:
// 			s, err := c.controller.GateStatus(ctx, addr)
// 			if err != nil {
// 				return err
// 			}

// 			fmt.Printf("%s\n", s.String())

// 			if check(s) {
// 				return nil
// 			}
// 			c.debugn("%s not complete...checking again", name)
// 		}
// 	}
// }

func (s *serviceLayer) debug(msg string, args ...interface{}) {
	if s.verbose {
		color.Green(fmt.Sprintf("[X] "+msg, args...))
	}
}

func (s *serviceLayer) warn(msg string, args ...interface{}) {
	if s.verbose {
		color.Magenta(fmt.Sprintf("[X] "+msg, args...))
	}
}
