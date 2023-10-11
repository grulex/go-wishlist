package httputil

import "net/http"

type HttpUseCase func(r *http.Request) HandleResult
type responseType string

const (
	ResponseTypeJson responseType = "json"
	ResponseTypeHtml responseType = "html"
	ResponseTypeJpeg responseType = "jpeg"
)

type HandleResult struct {
	Payload interface{}
	Type    responseType
	Error   *HandleError
}

func (h HandleResult) HasError() bool {
	return h.Error != nil
}
