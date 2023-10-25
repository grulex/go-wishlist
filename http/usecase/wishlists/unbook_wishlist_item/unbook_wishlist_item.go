package unbook_wishlist_item

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	GetWishlistItem(ctx context.Context, itemID wishlistPkg.ItemID) (*wishlistPkg.Item, error)
	UnBookItem(ctx context.Context, itemID wishlistPkg.ItemID) error
}

func MakeUnBookWishlistItemUsecase(wService wishlistService) httputil.HttpUseCase {
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
				},
			}
		}

		wishlist, err := wService.Get(r.Context(), wishlistPkg.ID(wishlistID))
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist",
				},
			}
		}
		isOwner := wishlist.UserID == auth.UserID

		if item.IsBookedBy != nil && *item.IsBookedBy != auth.UserID && !isOwner {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorForbidden,
					ErrorKey: "forbidden_booked_by_another_user",
					Message:  "item booked by another user",
					Err:      nil,
				},
			}
		}

		if err := wService.UnBookItem(r.Context(), itemID); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error unBooking wishlist item",
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
