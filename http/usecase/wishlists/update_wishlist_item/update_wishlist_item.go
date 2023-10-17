package update_wishlist_item

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	"github.com/grulex/go-wishlist/http/usecase/wishlists"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type requestJson struct {
	Product            types.Product `json:"product"`
	IsBookingAvailable bool          `json:"is_booking_available,omitempty"`
}

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	SetBookingAvailabilityForItem(ctx context.Context, itemID wishlistPkg.ItemID, isAvailable bool) error
	GetWishlistItem(ctx context.Context, itemID wishlistPkg.ItemID) (*wishlistPkg.Item, error)
}

type productService interface {
	Get(ctx context.Context, id productPkg.ID) (*productPkg.Product, error)
	Update(ctx context.Context, product *productPkg.Product) error
}

func MakeUpdateWishlistItemUsecase(wService wishlistService, pService productService) httputil.HttpUseCase {
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
		wishlistID, hasWishlistID := vars["id"]
		productID, hasProductID := vars["productId"]
		if !hasWishlistID || !hasProductID {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  "incorrect path parameter",
					Err:      nil,
				},
			}
		}

		handleResult, valid := wishlists.IsValidWishlistAccess(r.Context(), wService, wishlistID, auth)
		if !valid {
			return handleResult
		}

		jsonRequest := requestJson{}
		if err := json.NewDecoder(r.Body).Decode(&jsonRequest); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorBadData,
					Message: "Error decoding json request",
					Err:     err,
				},
			}
		}

		itemID := wishlistPkg.ItemID{
			WishlistID: wishlistPkg.ID(wishlistID),
			ProductID:  productPkg.ID(productID),
		}

		item, err := wService.GetWishlistItem(r.Context(), itemID)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist item",
					Err:     err,
				},
			}
		}
		item.IsBookingAvailable = jsonRequest.IsBookingAvailable
		if err := wService.SetBookingAvailabilityForItem(r.Context(), itemID, jsonRequest.IsBookingAvailable); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error setting booking availability for item",
					Err:     err,
				},
			}
		}

		product, err := pService.Get(r.Context(), productPkg.ID(productID))
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting product",
					Err:     err,
				},
			}
		}
		product.Title = jsonRequest.Product.Title
		product.Description = jsonRequest.Product.Description
		product.Url = jsonRequest.Product.Url
		product.Price = jsonRequest.Product.PriceFrom
		if jsonRequest.Product.Image != nil {
			product.ImageID = &jsonRequest.Product.Image.ID
		}
		if err := pService.Update(r.Context(), product); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error updating product",
					Err:     err,
				},
			}
		}

		return httputil.HandleResult{}
	}
}
