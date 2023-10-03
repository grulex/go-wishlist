package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

func ResponseWrapper(f HttpUseCase) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		result := f(r)
		if result.HasError() {
			responseError(result.Error, w)
			return
		}

		responseOk(result, w)
	}

	return handler
}

func responseError(handleError *HandleError, w http.ResponseWriter) {
	if handleError.Type == ErrorInternal {
		log.Println("Handler Error: ", string(handleError.JsonEncode()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	errorResp := struct {
		Error *HandleError `json:"error"`
	}{handleError}
	responseJson, err := json.Marshal(errorResp)
	if err != nil {
		log.Println("can't encode json", err)
		http.Error(w, "can't encode json response error", handleError.GetHttpStatus())
		return
	}
	w.WriteHeader(handleError.GetHttpStatus())
	w.Header().Set("Content-Type", "application/json")
	if n, err := w.Write(responseJson); err != nil {
		log.Println("error writing response", err, "bytesWritten", n)
	}
}

func responseOk(result HandleResult, w http.ResponseWriter) {
	if result.Type == ResponseTypeJson {
		responseJson, err := json.Marshal(result.Payload)
		if err != nil {
			log.Println("can't encode json", err)
			http.Error(w, "can't encode json response error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if n, err := w.Write(responseJson); err != nil {
			log.Println("error writing response", "err", err, "bytesWritten", n)
		}
		return
	}

	log.Println("unknown response type", result.Type)
	http.Error(w, "unknown response type", http.StatusInternalServerError)
}
