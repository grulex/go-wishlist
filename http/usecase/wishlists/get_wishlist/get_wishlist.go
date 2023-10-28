package get_wishlist

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type wishlistService interface {
	Get(ctx context.Context, id wishlistPkg.ID) (*wishlistPkg.Wishlist, error)
}

type subscribeService interface {
	Get(ctx context.Context, userID userPkg.ID, wishlistID wishlistPkg.ID) (*subscribePkg.Subscribe, error)
}

type imageService interface {
	Get(ctx context.Context, id imagePkg.ID) (*imagePkg.Image, error)
}

func MakeGetWishlistUsecase(sService subscribeService, wService wishlistService, iService imageService) httputil.HttpUseCase {
	return func(r *http.Request) httputil.HandleResult {
		var currentUserID *userPkg.ID
		auth, ok := authPkg.FromContext(r.Context())
		if ok {
			currentUserID = &auth.UserID
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

		wishlist, err := wService.Get(r.Context(), wishlistPkg.ID(id))
		if err != nil && !errors.Is(err, wishlistPkg.ErrNotFound) {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlist",
					Err:     err,
				},
			}
		}

		if wishlist == nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:     httputil.ErrorNotFound,
					ErrorKey: "not_found",
					Message:  fmt.Sprintf("wishlist with id %s not found", id),
					Err:      err,
				},
			}
		}

		var subscribe *subscribePkg.Subscribe
		if currentUserID != nil {
			subscribe, err = sService.Get(r.Context(), *currentUserID, wishlist.ID)
			if err != nil && !errors.Is(err, subscribePkg.ErrNotFound) {
				return httputil.HandleResult{
					Error: &httputil.HandleError{
						Type:    httputil.ErrorInternal,
						Message: "Error getting subscribe",
						Err:     err,
					},
				}
			}
		}

		var avatarAnswer *types.Image
		if wishlist.Avatar != nil {
			avatar, err := iService.Get(r.Context(), *wishlist.Avatar)
			if err != nil {
				return httputil.HandleResult{
					Error: &httputil.HandleError{
						Type:    httputil.ErrorInternal,
						Message: "Error getting avatar",
						Err:     err,
					},
				}
			}
			avatarAnswer = &types.Image{
				ID:   *wishlist.Avatar,
				Link: usecase.GetFileUrl(r, avatar.FileLink),
			}
		}

		payload := struct {
			Wishlist     types.Wishlist `json:"wishlist"`
			IsSubscribed bool           `json:"is_subscribed"`
		}{
			Wishlist: types.Wishlist{
				ID:           wishlist.ID,
				Title:        wishlist.Title,
				Description:  wishlist.Description,
				IsDefault:    wishlist.IsDefault,
				Avatar:       avatarAnswer,
				IsMyWishlist: currentUserID != nil && *currentUserID == wishlist.UserID,
			},
			IsSubscribed: subscribe != nil,
		}

		return httputil.HandleResult{
			Payload: payload,
			Type:    httputil.ResponseTypeJson,
		}
	}
}
