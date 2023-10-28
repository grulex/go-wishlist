package wishlists

import (
	"context"
	"errors"
	"github.com/grulex/go-wishlist/http/httputil"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
}

func IsValidWishlistAccess(ctx context.Context, wService wishlistService, wishlistID string, auth *authPkg.Auth) (httputil.HandleResult, bool) {
	wishlist, err := wService.Get(ctx, wishlistPkg.ID(wishlistID))
	if err != nil && !errors.Is(err, wishlistPkg.ErrNotFound) {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorInternal,
				Message: "Error getting wishlist",
				Err:     err,
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
	if wishlist.UserID != auth.UserID {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:     httputil.ErrorForbidden,
				ErrorKey: "forbidden",
				Message:  "you haven't access to change this wishlist",
				Err:      nil,
			},
		}, false
	}

	return httputil.HandleResult{}, true
}
