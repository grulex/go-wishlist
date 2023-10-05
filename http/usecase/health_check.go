package usecase

import (
	"github.com/grulex/go-wishlist/http/httputil"
	"net/http"
)

func MakeHealthCheckUsecase() httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		payload := struct {
			Alive bool `json:"alive"`
		}{true}

		return httputil.HandleResult{
			Payload: payload,
			Type:    httputil.ResponseTypeJson,
		}
	}
}
