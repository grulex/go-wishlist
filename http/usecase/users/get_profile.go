package users

import (
	"context"
	"github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/usecase/types"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	subscribePkg "github.com/grulex/go-wishlist/pkg/subscribe"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"net/http"
)

type subscribeService interface {
	GetByUser(ctx context.Context, userID userPkg.ID) ([]*subscribePkg.Subscribe, error)
}

type wishlistService interface {
	GetByUserID(ctx context.Context, userID userPkg.ID) ([]*wishlistPkg.Wishlist, error)
}

func MakeGetProfileUsecase(subscribesService subscribeService, wService wishlistService) httputil.HttpUseCase {
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
			avatarAnswer = &types.Image{
				ID:   defaultWishlist.Avatar.ID,
				Link: "", // todo
			}
		}

		payload := struct {
			User            types.User        `json:"user"`
			DefaultWishlist types.Wishlist    `json:"defaultWishlist"`
			Subscribes      []types.Subscribe `json:"subscribes"`
		}{
			User: types.User{
				ID: auth.UserID,
			},
			DefaultWishlist: types.Wishlist{
				ID:          defaultWishlist.ID,
				IsDefault:   defaultWishlist.IsDefault,
				Title:       defaultWishlist.Title,
				Avatar:      avatarAnswer,
				Description: defaultWishlist.Description,
			},
			Subscribes: subscribeAnswer,
		}

		return httputil.HandleResult{
			Payload: payload,
			Type:    httputil.ResponseTypeJson,
		}
	}
}
