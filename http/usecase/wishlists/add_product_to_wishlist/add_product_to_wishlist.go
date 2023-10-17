package add_product_to_wishlist

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

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
	AddWishlistItem(ctx context.Context, item *wishlistPkg.Item) error
}

type productService interface {
	Create(ctx context.Context, product *productPkg.Product) error
	Get(ctx context.Context, id productPkg.ID) (*productPkg.Product, error)
}

type requestJson struct {
	Product            types.Product `json:"product"`
	IsBookingAvailable bool          `json:"is_booking_available,omitempty"`
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

		handleResult, valid := wishlists.IsValidWishlistAccess(r.Context(), wService, wishlistID, auth)
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

		product := &productPkg.Product{
			Title:       request.Product.Title,
			Price:       request.Product.PriceFrom,
			Description: request.Product.Description,
			Url:         request.Product.Url,
			ImageID:     nil, // todo
		}

		// for "copy to my wishlist" feature
		if request.Product.ID != nil && *request.Product.ID != "" {
			// todo is cloned from field
			p, err := pService.Get(r.Context(), *request.Product.ID)
			if err != nil {
				return httputil.HandleResult{
					Error: &httputil.HandleError{
						Type:    httputil.ErrorInternal,
						Message: "Error getting product",
					},
				}
			}

			product.Title = p.Title
			product.Price = p.Price
			product.Description = p.Description
			product.Url = p.Url
			product.ImageID = p.ImageID
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
