package update_wishlist

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	"github.com/grulex/go-wishlist/http/usecase/wishlists"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	filePkg "github.com/grulex/go-wishlist/pkg/file"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"io"
	"log"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	Update(ctx context.Context, wishlist *wishlistPkg.Wishlist) error
}

type requestJson struct {
	Wishlist types.Wishlist `json:"wishlist"`
}

type fileService interface {
	UploadPhoto(ctx context.Context, reader io.Reader) (filePkg.Link, error)
}

type imageService interface {
	Create(ctx context.Context, image *imagePkg.Image) error
}

func MakeUpdateWishlistUsecase(wService wishlistService, fService fileService, iService imageService) httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		auth, ok := authPkg.FromContext(r.Context())
		if !ok {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Message: "Unauthorized",
					Type:    httputil.ErrorBadAuth,
				},
			}
		}

		vars := mux.Vars(r)
		wishlistID, ok := vars["id"]
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

		wishlist, err := wService.Get(r.Context(), wishlistPkg.ID(wishlistID))
		if err != nil && !errors.Is(err, wishlistPkg.ErrNotFound) {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist",
				},
			}
		}

		handleResult, valid := wishlists.IsValidWishlistAccess(r.Context(), wService, wishlistID, auth)
		if !valid {
			return handleResult
		}

		request := requestJson{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorBadData,
					Message: "invalid json body",
					Err:     err,
				},
			}
		}

		if request.Wishlist.Avatar != nil && request.Wishlist.Avatar.ID == "" {
			newImage, result := wishlists.UploadBase64Image(r.Context(), fService, iService, request.Wishlist.Avatar.Src)
			if result.Error != nil {
				log.Printf("Error uploading image: %v, %+v\n", result.Error, request.Wishlist.Avatar)
				return result
			}
			wishlist.Avatar = &newImage.ID
		}

		wishlist.Title = request.Wishlist.Title
		wishlist.Description = request.Wishlist.Description
		wishlist.IsDefault = request.Wishlist.IsDefault
		err = wService.Update(r.Context(), wishlist)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error updating wishlist",
					Err:     err,
				},
			}
		}

		return httputil.HandleResult{}
	}
}
