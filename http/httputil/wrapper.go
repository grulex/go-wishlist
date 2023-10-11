package httputil

import (
	"encoding/json"
	"html/template"
	"io"
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
		log.Println("Handler Error: ", string(handleError.JsonEncode()), "err", handleError.Err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	errorResp := struct {
		Error *HandleError `json:"error"`
	}{handleError}
	responseJson, err := json.Marshal(errorResp)
	if err != nil {
		log.Println("can't encode json", err, "error", handleError.Err.Error())
		http.Error(w, "can't encode json response error", handleError.GetHttpStatus())
		return
	}
	w.WriteHeader(handleError.GetHttpStatus())
	w.Header().Set("Content-Type", "application/json")
	if n, err := w.Write(responseJson); err != nil {
		log.Println("error writing response", err, "bytesWritten", n, "error", handleError.Err.Error())
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

	if result.Type == ResponseTypeHtml {
		tmpl := result.Payload.(*template.Template)
		if err := tmpl.Execute(w, nil); err != nil {
			log.Println("error executing template", err)
			http.Error(w, "error executing template", http.StatusInternalServerError)
		}
		return
	}

	if result.Type == ResponseTypeJpeg {
		reader := result.Payload.(io.ReadCloser)
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		if n, err := io.Copy(w, reader); err != nil {
			log.Println("error writing response", "err", err, "bytesWritten", n)
		}
		_ = reader.Close()
		return
	}

	responseJson, _ := json.Marshal(struct{}{})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if n, err := w.Write(responseJson); err != nil {
		log.Println("error writing response", "err", err, "bytesWritten", n)
	}
}
