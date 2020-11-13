package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type Service interface {
	PushButtonOpen(ctx context.Context) error
	PushButtonClose(ctx context.Context) error
}

type (
	PushButtonRequest  struct{}
	PushButtonResponse struct {
		Ok    bool
		Error string
	}
)

func New(svc Service) (http.Handler, error) {
	r := mux.NewRouter()

	r.HandleFunc("/push-button-open", MakePushButtonOpenHandler(svc)).Methods("POST")
	r.HandleFunc("/push-button-close", MakePushButtonCloseHandler(svc)).Methods("POST")

	return r, nil
}

func MakePushButtonOpenHandler(svc Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("request Push Button Open")

		ctx := context.Background()
		err := svc.PushButtonOpen(ctx)

		var res *PushButtonResponse
		if err != nil {
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

func MakePushButtonCloseHandler(svc Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Println("request Push Button Close")

		ctx := context.Background()
		err := svc.PushButtonClose(ctx)

		var res *PushButtonResponse
		if err != nil {
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

func encodeJsonResponse(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
