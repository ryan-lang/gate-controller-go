package ops

import (
	"context"
	"gate/control/messages"
	"time"
)

type Operation interface {
	ID() string
	Timeout() time.Duration
	ControlRequest() *messages.GateControlRequest
	CheckDone(context.Context, *messages.GateStatusResponse) bool
	CheckInProgress(context.Context, *messages.GateStatusResponse) bool
	CheckFault(context.Context, *messages.GateStatusResponse) error
}
