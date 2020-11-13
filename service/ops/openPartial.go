package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type OpenPartialOp struct {
	address int
}

func NewOpenPartialOp(addr int) *OpenPartialOp {
	return &OpenPartialOp{
		address: addr,
	}
}

func (o *OpenPartialOp) ID() string {
	return "OpenPartial"
}

func (o *OpenPartialOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *OpenPartialOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:     o.address,
		OpenPartial: true,
	}
}

func (o *OpenPartialOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.PartialOpenLimit
}

func (o *OpenPartialOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return false
}

func (o *OpenPartialOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
