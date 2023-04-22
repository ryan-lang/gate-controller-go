package rpc

import (
	"context"
	msgs "gate/control/messages"
	ops "gate/service/ops"
	"net/rpc"
)

type GateService struct {
	service Service
}

type Service interface {
	Listen(
		ctx context.Context,
		stCh chan *msgs.GateStatusResponse,
		fCh chan *msgs.GateFaultResponse,
		erCh chan error,
	) int
	PushButtonOpen(ctx context.Context) (int32, error)
	PushButtonClose(ctx context.Context) error
	Exec(ctx context.Context, op ops.Operation) error
}

type (
	ListenRequest struct {
		StatusChannel chan *msgs.GateStatusResponse
		FaultChannel  chan *msgs.GateFaultResponse
		ErrorChannel  chan error
	}
	ListenResponse struct {
		Ok    bool
		Error error
	}
	PushButtonRequest  struct{}
	PushButtonResponse struct {
		Ok          bool
		Error       error
		GateEventID int32
	}
	ExecRequest struct {
		Operation ops.Operation
	}
	ExecResponse struct {
		Ok    bool
		Error error
	}
)

func HandleHTTP(service Service) error {
	tp := &GateService{
		service: service,
	}
	err := rpc.Register(tp)
	if err != nil {
		return err
	}
	rpc.HandleHTTP()
	return nil
}

func (t *GateService) Listen(req *ListenRequest, res *ListenResponse) error {
	ctx := context.Background()
	t.service.Listen(ctx, req.StatusChannel, req.FaultChannel, req.ErrorChannel)
	res.Ok = true
	return nil
}

func (t *GateService) PushButtonOpen(req *PushButtonRequest, res *PushButtonResponse) error {
	ctx := context.Background()
	gateEventID, err := t.service.PushButtonOpen(ctx)

	res.GateEventID = gateEventID
	res.Error = err
	if err != nil {
		res.Ok = false
	}

	return nil
}

func (t *GateService) PushButtonClose(req *PushButtonRequest, res *PushButtonResponse) error {
	ctx := context.Background()
	err := t.service.PushButtonClose(ctx)

	res.Error = err
	if err != nil {
		res.Ok = false
	}

	return nil
}

func (t *GateService) Exec(req *ExecRequest, res *ExecResponse) error {
	ctx := context.Background()
	err := t.service.Exec(ctx, req.Operation)

	res.Error = err
	if err != nil {
		res.Ok = false
	}

	return nil
}
