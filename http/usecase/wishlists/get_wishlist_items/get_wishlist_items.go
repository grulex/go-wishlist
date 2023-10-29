package get_wishlist_items

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	productPkg "github.com/grulex/go-wishlist/pkg/product"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
	"strconv"
)

type wishlistService interface {
	GetWishlistItems(ctx context.Context, wishlistID wishlistPkg.ID, limit, offset uint) ([]*wishlistPkg.Item, bool, error)
}

type productService interface {
	GetMany(ctx context.Context, ids []productPkg.ID) ([]*productPkg.Product, error)
}

type imageService interface {
	GetMany(ctx context.Context, ids []imagePkg.ID) ([]*imagePkg.Image, error)
}

func MakeGetWishlistItemsUsecase(wService wishlistService, productService productService, iService imageService) httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		var currentUserID *userPkg.ID
		auth, ok := authPkg.FromContext(r.Context())
		if ok {
			currentUserID = &auth.UserID
		}
		var err error
		var limit, offset int64
		if len(r.URL.Query().Get("limit")) == 0 {
			limit, err = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
			if err != nil {
				limit = 100
			}
		}
		if len(r.URL.Query().Get("offset")) == 0 {
			offset, err = strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
			if err != nil {
				offset = 0
			}
		}

		vars := mux.Vars(r)
		id, ok := vars["id"]
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

		items, hasMore, err := wService.GetWishlistItems(r.Context(), wishlistPkg.ID(id), uint(limit), uint(offset))
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist items",
					Err:     err,
				},
			}
		}

		ids := make([]productPkg.ID, len(items))
		for i, item := range items {
			ids[i] = item.ID.ProductID
		}

		products, err := productService.GetMany(r.Context(), ids)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting products",
					Err:     err,
				},
			}
		}

		productsMap := make(map[productPkg.ID]*productPkg.Product, len(products))
		imageIds := make([]imagePkg.ID, 0, len(products))
		for _, product := range products {
			productsMap[product.ID] = product
			if product.ImageID != nil {
				imageIds = append(imageIds, *product.ImageID)
			}
		}

		images, err := iService.GetMany(r.Context(), imageIds)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting images",
					Err:     err,
				},
			}
		}

		imagesMap := make(map[imagePkg.ID]*imagePkg.Image, len(images))
		for _, image := range images {
			imagesMap[image.ID] = image
		}

		var resultItems []types.Item
		for _, item := range items {
			product := productsMap[item.ID.ProductID]
			isBookedByCurrentUser := currentUserID != nil &&
				item.IsBookedBy != nil &&
				*item.IsBookedBy == *currentUserID
			var resImage *types.Image
			if productsMap[item.ID.ProductID].ImageID != nil {
				image := imagesMap[*product.ImageID]

				sizes := make([]types.ImageSize, len(image.Sizes))
				for i, s := range image.Sizes {
					sizes[i] = types.ImageSize{
						Width:  s.Width,
						Height: s.Height,
						Link:   usecase.GetFileUrl(r, s.FileLink),
					}
				}

				resImage = &types.Image{
					ID:    *product.ImageID,
					Link:  usecase.GetFileUrl(r, image.FileLink),
					Sizes: sizes,
				}
			}
			resultItems = append(resultItems, types.Item{
				ID:                    item.ID,
				IsBookingAvailable:    item.IsBookingAvailable,
				IsBookedByCurrentUser: isBookedByCurrentUser,
				IsBooked:              item.IsBookedBy != nil,
				Product: types.Product{
					ID:          &item.ID.ProductID,
					Title:       productsMap[item.ID.ProductID].Title,
					PriceFrom:   productsMap[item.ID.ProductID].Price,
					Description: productsMap[item.ID.ProductID].Description,
					Url:         productsMap[item.ID.ProductID].Url,
					Image:       resImage,
				},
			})
		}

		payload := struct {
			Items   []types.Item `json:"items"`
			HasMore bool         `json:"has_more"`
		}{
			Items:   resultItems,
			HasMore: hasMore,
		}

		return httputil.HandleResult{
			Payload: payload,
			Type:    httputil.ResponseTypeJson,
		}
	}
}
