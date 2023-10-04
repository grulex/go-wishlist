package httputil

import (
	"encoding/json"
	"net/http"
)

type errorType string

type HandleError struct {
	Type     errorType `json:"type"`
	ErrorKey string    `json:"error_key"`
	Message  string    `json:"message"`
	Err      error     `json:"-"`
}

func (h *HandleError) JsonEncode() []byte {
	j, _ := json.Marshal(h)
	return j
}

func (h *HandleError) GetHttpStatus() int {
	switch h.Type {
	case ErrorBadData:
		return http.StatusBadRequest
	case ErrorNotFound:
		return http.StatusNotFound
	case ErrorBadAuth:
		return http.StatusUnauthorized
	case ErrorForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

const (
	ErrorBadData   errorType = "bad_data"
	ErrorNotFound  errorType = "not_found"
	ErrorBadAuth   errorType = "bad_auth"
	ErrorForbidden errorType = "no_permission"
	ErrorInternal  errorType = "internal"
)
