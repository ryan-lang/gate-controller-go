package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type PushButtonCloseOp struct {
	address int
}

func NewPushButtonCloseOp(addr int) *PushButtonCloseOp {
	return &PushButtonCloseOp{
		address: addr,
	}
}

func (o *PushButtonCloseOp) ID() string {
	return "PushButtonClose"
}

func (o *PushButtonCloseOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *PushButtonCloseOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:         o.address,
		PushButtonClose: true,
	}
}

func (o *PushButtonCloseOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.CloseLimit
}

func (o *PushButtonCloseOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.LastCommandStatus == msgs.STATUS_CLOSE_INPROGRESS
}

func (o *PushButtonCloseOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	if s.LastCommandStatus == msgs.STATUS_STOPPED {
		return &ErrOpStopped{}
	}

	return nil
}
