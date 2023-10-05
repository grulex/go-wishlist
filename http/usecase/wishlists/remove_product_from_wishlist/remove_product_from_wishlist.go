package remove_product_from_wishlist

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	GetWishlistItem(ctx context.Context, itemID wishlistPkg.ItemID) (*wishlistPkg.Item, error)
	RemoveItem(ctx context.Context, item wishlistPkg.ItemID) error
}

func MakeRemoveProductFromWishlistUsecase(wService wishlistService) httputil.HttpUseCase {
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

		handleResult, valid := isValidWishlistAccess(r, wService, wishlistID, auth.UserID)
		if !valid {
			return handleResult
		}

		productId, ok := vars["productId"]
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

		itemID := wishlistPkg.ItemID{
			WishlistID: wishlistPkg.ID(wishlistID),
			ProductID:  productPkg.ID(productId),
		}

		if err := wService.RemoveItem(r.Context(), itemID); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error removing item from wishlist",
				},
			}
		}

		return httputil.HandleResult{}
	}
}

func isValidWishlistAccess(r *http.Request, wService wishlistService, wishlistID string, currentUserID userPkg.ID) (httputil.HandleResult, bool) {
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
