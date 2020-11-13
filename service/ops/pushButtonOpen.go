package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type PushButtonOpenOp struct {
	address int
}

func NewPushButtonOpenOp(addr int) *PushButtonOpenOp {
	return &PushButtonOpenOp{
		address: addr,
	}
}

func (o *PushButtonOpenOp) ID() string {
	return "PushButtonOpen"
}

func (o *PushButtonOpenOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *PushButtonOpenOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:        o.address,
		PushButtonOpen: true,
	}
}

func (o *PushButtonOpenOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.OpenLimit
}

func (o *PushButtonOpenOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.LastCommandStatus == msgs.STATUS_OPEN_INPROGRESS
}

func (o *PushButtonOpenOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	if s.LastCommandStatus == msgs.STATUS_STOPPED {
		return &ErrOpStopped{}
	}

	return nil
}
