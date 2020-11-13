package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type OpenInterlockOp struct {
	address int
}

func NewOpenInterlockOp(addr int) *OpenInterlockOp {
	return &OpenInterlockOp{
		address: addr,
	}
}

func (o *OpenInterlockOp) ID() string {
	return "OpenInterlock"
}

func (o *OpenInterlockOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *OpenInterlockOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:       o.address,
		OpenInterlock: true,
	}
}

func (o *OpenInterlockOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.PartialOpenLimit
}

func (o *OpenInterlockOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return false
}

func (o *OpenInterlockOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
