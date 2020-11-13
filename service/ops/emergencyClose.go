package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type EmergencyCloseOp struct {
	address int
}

func NewEmergencyCloseOp(addr int) *EmergencyCloseOp {
	return &EmergencyCloseOp{
		address: addr,
	}
}

func (o *EmergencyCloseOp) ID() string {
	return "EmergencyClose"
}

func (o *EmergencyCloseOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *EmergencyCloseOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:        o.address,
		EmergencyClose: true,
	}
}

func (o *EmergencyCloseOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.CloseLimit
}

func (o *EmergencyCloseOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.LastCommandStatus == msgs.STATUS_CLOSE_INPROGRESS
}

func (o *EmergencyCloseOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
