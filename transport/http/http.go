package http

import (
	"context"
	"encoding/json"
	"fmt"
	msgs "gate/control/messages"
	"gate/service"
	"net/http"

	"github.com/gorilla/mux"
)

type Service interface {
	PushButtonOpen(ctx context.Context) (int32, error)
	PushButtonClose(ctx context.Context) error
	LastStatus() *msgs.GateStatusResponse
	Metrics() *service.ServiceMetrics
}

type (
	PushButtonRequest  struct{}
	PushButtonResponse struct {
		GateEventID int32
		Ok          bool
		Error       string
	}

	StateResponse struct {
		Status  *msgs.GateStatusResponse
		Metrics *service.ServiceMetrics
	}
)

func New(svc Service) (http.Handler, error) {
	r := mux.NewRouter()

	r.HandleFunc("/state", MakeStateHandler(svc)).Methods("GET")
	r.HandleFunc("/push-button-open", MakePushButtonOpenHandler(svc)).Methods("POST")
	r.HandleFunc("/push-button-close", MakePushButtonCloseHandler(svc)).Methods("POST")
	r.HandleFunc("/control/push-button-open", MakePushButtonOpenHandler(svc)).Methods("POST")
	r.HandleFunc("/control/push-button-close", MakePushButtonCloseHandler(svc)).Methods("POST")

	return r, nil
}

func MakePushButtonOpenHandler(svc Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("request Push Button Open")

		ctx := context.Background()
		gateEventID, err := svc.PushButtonOpen(ctx)

		var res *PushButtonResponse
		if err != nil {
			fmt.Printf("push button open failed: %s\n", err.Error())
			res = &PushButtonResponse{
				Ok:          false,
				Error:       err.Error(),
				GateEventID: gateEventID,
			}
		} else {
			res = &PushButtonResponse{
				Ok:          true,
				GateEventID: gateEventID,
			}
		}

		encodeJsonResponse(w, res)
	}
}

func MakePushButtonCloseHandler(svc Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("request Push Button Close")

		ctx := context.Background()
		err := svc.PushButtonClose(ctx)

		var res *PushButtonResponse
		if err != nil {
			fmt.Printf("push button close failed: %s\n", err.Error())
			res = &PushButtonResponse{
				Ok:    false,
				Error: err.Error(),
			}
		} else {
			res = &PushButtonResponse{
				Ok: true,
			}
		}

		encodeJsonResponse(w, res)
	}
}

func MakeStateHandler(svc Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("request state")
		res := &StateResponse{
			Status:  svc.LastStatus(),
			Metrics: svc.Metrics(),
		}
		encodeJsonResponse(w, res)
	}
}

func encodeJsonResponse(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
