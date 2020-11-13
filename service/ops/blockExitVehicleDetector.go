package ops

import (
	"context"
	msgs "gate/control/messages"
	"time"
)

type BlockExitVehicleDetectorOp struct {
	address int
}

func NewBlockExitVehicleDetectorOp(addr int) *BlockExitVehicleDetectorOp {
	return &BlockExitVehicleDetectorOp{
		address: addr,
	}
}

func (o *BlockExitVehicleDetectorOp) ID() string {
	return "BlockExitVehicleDetector"
}

func (o *BlockExitVehicleDetectorOp) Timeout() time.Duration {
	return time.Second * 10
}

func (o *BlockExitVehicleDetectorOp) ControlRequest() *msgs.GateControlRequest {
	return &msgs.GateControlRequest{
		Address:                  o.address,
		BlockExitVehicleDetector: true,
	}
}

func (o *BlockExitVehicleDetectorOp) CheckDone(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return s.PartialOpenLimit
}

func (o *BlockExitVehicleDetectorOp) CheckInProgress(ctx context.Context, s *msgs.GateStatusResponse) bool {
	return false
}

func (o *BlockExitVehicleDetectorOp) CheckFault(ctx context.Context, s *msgs.GateStatusResponse) error {
	return nil
}
