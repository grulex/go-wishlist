package book_wishlist_item

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
	BookItem(ctx context.Context, itemID wishlistPkg.ItemID, userID userPkg.ID) error
}

func MakeBookWishlistItemUsecase(wService wishlistService) httputil.HttpUseCase {
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

		handleResult, valid := isValidWishlist(r, wService, wishlistID)
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

		item, err := wService.GetWishlistItem(r.Context(), itemID)
		if err != nil && !errors.Is(err, wishlistPkg.ErrItemNotFound) {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist item",
					Err:     err,
				},
			}
		}
		if item.IsBookedBy != nil && item.IsBookedBy != &auth.UserID {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorForbidden,
					ErrorKey: "forbidden_booked_by_another_user",
					Message:  "item already booked by another user",
					Err:      nil,
				},
			}
		}

		if err := wService.BookItem(r.Context(), itemID, auth.UserID); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error booking wishlist item",
					Err:     err,
				},
			}
		}

		return httputil.HandleResult{}
	}
}

func isValidWishlist(r *http.Request, wService wishlistService, wishlistID string) (httputil.HandleResult, bool) {
	wishlist, err := wService.Get(r.Context(), wishlistPkg.ID(wishlistID))
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

	return httputil.HandleResult{}, true
}
