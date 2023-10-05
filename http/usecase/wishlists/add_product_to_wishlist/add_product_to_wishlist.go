package add_product_to_wishlist

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	AddWishlistItem(ctx context.Context, item *wishlistPkg.Item) error
}

type productService interface {
	Create(ctx context.Context, product *productPkg.Product) error
}

type requestJson struct {
	Product            types.Product `json:"product"`
	IsBookingAvailable bool          `json:"is_booking_available"`
}

func MakeAddProductToWishlistUsecase(wService wishlistService, pService productService) httputil.HttpUseCase {
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

		handleResult, valid := isValidWishlistAccess(r, wService, wishlistID, auth)
		if !valid {
			return handleResult
		}

		request := requestJson{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorBadData,
					ErrorKey: "bad_item_json",
					Message:  "incorrect data in request body",
					Err:      nil,
				},
			}
		}

		if request.Product.ID != nil {
			return addItemToWishlist(r.Context(), wishlistID, *request.Product.ID, request.IsBookingAvailable, wService)
		}

		product := &productPkg.Product{
			Title:       request.Product.Title,
			PriceFrom:   request.Product.PriceFrom,
			PriceTo:     request.Product.PriceTo,
			Description: request.Product.Description,
			Url:         request.Product.Url,
			ImageID:     nil, // todo
		}
		if err := pService.Create(r.Context(), product); err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error creating product",
				},
			}
		}

		return addItemToWishlist(r.Context(), wishlistID, product.ID, request.IsBookingAvailable, wService)
	}
}

func isValidWishlistAccess(r *http.Request, wService wishlistService, wishlistID string, auth *authPkg.Auth) (httputil.HandleResult, bool) {
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
	if wishlist.UserID != auth.UserID {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:     httputil.ErrorForbidden,
				ErrorKey: "forbidden",
				Message:  "you can't add product to wishlist of another user",
				Err:      nil,
			},
		}, false
	}

	return httputil.HandleResult{}, true
}

func addItemToWishlist(
	ctx context.Context,
	wishlistID string,
	productID productPkg.ID,
	isBookingAvailable bool,
	wService wishlistService,
) httputil.HandleResult {
	item := &wishlistPkg.Item{
		ID: wishlistPkg.ItemID{
			WishlistID: wishlistPkg.ID(wishlistID),
			ProductID:  productID,
		},
		IsBookingAvailable: isBookingAvailable,
		IsBookedBy:         nil,
	}
	if err := wService.AddWishlistItem(ctx, item); err != nil {
		return httputil.HandleResult{
			Error: &httputil.HandleError{
				Type:    httputil.ErrorInternal,
				Message: "Error adding wishlist item",
			},
		}
	}

	return httputil.HandleResult{}
}
