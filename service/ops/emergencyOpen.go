package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type EmergencyOpenOp struct {
	address int
}

func NewEmergencyOpenOp(addr int) *EmergencyOpenOp {
	return &EmergencyOpenOp{
		address: addr,
	}
}

func (o *EmergencyOpenOp) ID() string {
	return "EmergencyOpen"
}

func (o *EmergencyOpenOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *EmergencyOpenOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:       o.address,
		EmergencyOpen: true,
	}
}

func (o *EmergencyOpenOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.OpenLimit
}

func (o *EmergencyOpenOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.LastCommandStatus == msgs.STATUS_OPEN_INPROGRESS
}

func (o *EmergencyOpenOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
