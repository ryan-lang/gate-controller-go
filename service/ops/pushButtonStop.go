package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type PushButtonStopOp struct {
	address int
}

func NewPushButtonStopOp(addr int) *PushButtonStopOp {
	return &PushButtonStopOp{
		address: addr,
	}
}

func (o *PushButtonStopOp) ID() string {
	return "PushButtonStop"
}

func (o *PushButtonStopOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *PushButtonStopOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:        o.address,
		PushButtonStop: true,
	}
}

func (o *PushButtonStopOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.LastCommandStatus == msgs.STATUS_STOPPED
}

func (o *PushButtonStopOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return false
}

func (o *PushButtonStopOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
