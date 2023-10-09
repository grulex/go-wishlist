package update_wishlist

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	Update(ctx context.Context, wishlist *wishlistPkg.Wishlist) error
}

type requestJson struct {
	Wishlist types.Wishlist `json:"wishlist"`
}

func MakeUpdateWishlistUsecase(wService wishlistService) httputil.HttpUseCase {
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

		handleResult, valid := isValidWishlistAccess(r, wishlist, wishlistID, auth.UserID)
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

func isValidWishlistAccess(r *http.Request, wishlist *wishlistPkg.Wishlist, wishlistID string, currentUserID userPkg.ID) (httputil.HandleResult, bool) {
	if wishlist == nil {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:     httputil.ErrorNotFound,
				ErrorKey: "not_found",
				Message:  "incorrect path parameter",
				Err:      nil,
			},
		}, false
	}
	if wishlist.UserID != currentUserID {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:     httputil.ErrorForbidden,
				ErrorKey: "forbidden",
				Message:  "you can't remove product from wishlist of another user",
				Err:      nil,
			},
		}, false
	}

	return httputil.HandleResult{}, true
}
