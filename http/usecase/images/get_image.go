package images

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/pkg/file"
	"io"
	"net/http"
)

type fileService interface {
	Download(ctx context.Context, link file.Link) (io.ReadCloser, error)
}

func MakeGetImageFileHandler(fileService fileService) httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		vars := mux.Vars(r)
		linkBase64, ok := vars["link_base64"]
		if !ok {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}
		link, err := file.NewLinkFromBase64(linkBase64)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}

		readCloser, err := fileService.Download(r.Context(), link)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "image not found",
					Err:      err,
				},
			}
		}
		return httputil.HandleResult{
			Payload: readCloser,
			Type:    httputil.ResponseTypeJpeg,
		}
	}
}
