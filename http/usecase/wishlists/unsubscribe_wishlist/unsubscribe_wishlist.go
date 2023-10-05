package unsubscribe_wishlist

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
}

type subscribeService interface {
	Unsubscribe(ctx context.Context, userID userPkg.ID, wishlistID wishlistPkg.ID) error
}

func MakeUnSubscribeWishlistUsecase(wService wishlistService, sService subscribeService) httputil.HttpUseCase {
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

		handleResult, valid := isValidWishlistAccess(r, wService, wishlistID)
		if !valid {
			return handleResult
		}

		err := sService.Unsubscribe(r.Context(), auth.UserID, wishlistPkg.ID(wishlistID))
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error subscribing to wishlist",
				},
			}
		}

		return httputil.HandleResult{}
	}
}

func isValidWishlistAccess(r *http.Request, wService wishlistService, wishlistID string) (httputil.HandleResult, bool) {
	wishlist, err := wService.Get(r.Context(), wishlistPkg.ID(wishlistID))
	if err != nil && !errors.Is(err, wishlistPkg.ErrNotFound) {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorInternal,
				Message: "Error getting wishlist",
			},
		}, false
	}
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

	return httputil.HandleResult{}, true
}
