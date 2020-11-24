package publisher

import (
	"bytes"
	"encoding/json"
	"gate/control/messages"
	"net/http"
)

type httpPublisher struct {
	addr string
}

type errorRequest struct {
	Error string
}

func New(publishAddr string) *httpPublisher {
	return &httpPublisher{
		addr: publishAddr,
	}
}

func (p *httpPublisher) PublishStatus(status *messages.GateStatusResponse) error {
	data, _ := json.Marshal(status)
	res, err := http.Post(p.addr+"/status", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (p *httpPublisher) PublishFault(fault *messages.GateFaultResponse) error {
	data, _ := json.Marshal(fault)
	res, err := http.Post(p.addr+"/fault", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func (p *httpPublisher) PublishError(e error) error {
	req := &errorRequest{e.Error()}
	data, _ := json.Marshal(req)
	res, err := http.Post(p.addr+"/error", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}
