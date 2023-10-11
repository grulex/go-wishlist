package users

import (
	"context"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	imagePkg "github.com/grulex/go-wishlist/pkg/image"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type subscribeService interface {
	GetByUser(ctx context.Context, userID userPkg.ID) ([]*subscribePkg.Subscribe, error)
}

type wishlistService interface {
	GetByUserID(ctx context.Context, userID userPkg.ID) (wishlistPkg.Wishlists, error)
}

type imageService interface {
	Get(ctx context.Context, id imagePkg.ID) (*imagePkg.Image, error)
}

func MakeGetProfileUsecase(subscribesService subscribeService, wService wishlistService, iService imageService) httputil.HttpUseCase {
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

		wishlists, err := wService.GetByUserID(r.Context(), auth.UserID)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting wishlists",
				},
			}
		}
		var defaultWishlist *wishlistPkg.Wishlist
		for _, w := range wishlists {
			if w.IsDefault {
				defaultWishlist = w
				break
			}
		}

		subscribes, err := subscribesService.GetByUser(r.Context(), auth.UserID)
		if err != nil {
			return httputil.HandleResult{
				Error: &httputil.HandleError{
					Type:    httputil.ErrorInternal,
					Message: "Error getting subscribes",
				},
			}
		}

		subscribeAnswer := make([]types.Subscribe, len(subscribes))
		for i, s := range subscribes {
			subscribeAnswer[i] = types.Subscribe{
				ID: s.WishlistID,
			}
		}

		var avatarAnswer *types.Image
		if defaultWishlist.Avatar != nil {
			avatar, err := iService.Get(r.Context(), *defaultWishlist.Avatar)
			if err != nil {
				return httputil.HandleResult{
					Error: &httputil.HandleError{
						Type:    httputil.ErrorInternal,
						Message: "Error getting avatar",
					},
				}
			}
			avatarAnswer = &types.Image{
				ID:   *defaultWishlist.Avatar,
				Link: string(avatar.FileLink.ID),
			}
		}

		payload := struct {
			User            types.User        `json:"user"`
			DefaultWishlist types.Wishlist    `json:"default_wishlist"`
			Subscribes      []types.Subscribe `json:"subscribes"`
		}{
			User: types.User{
				ID: auth.UserID,
			},
			DefaultWishlist: types.Wishlist{
				ID:           defaultWishlist.ID,
				IsDefault:    defaultWishlist.IsDefault,
				Title:        defaultWishlist.Title,
				Avatar:       avatarAnswer,
				Description:  defaultWishlist.Description,
				IsMyWishlist: true,
			},
			Subscribes: subscribeAnswer,
		}

		return httputil.HandleResult{
			Payload: payload,
			Type:    httputil.ResponseTypeJson,
		}
	}
}
