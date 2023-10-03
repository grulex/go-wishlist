package middleware

import (
	"context"
	"github.com/gorilla/mux"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	httpPkg "net/http"
)

type authService interface {
	Get(ctx context.Context, method authPkg.Method, socialID authPkg.SocialID) (authPkg.Auth, error)
	Create(ctx context.Context, auth authPkg.Auth) error
}

type userService interface {
	Create(ctx context.Context, user userPkg.User) error
}

func NewAuthMiddleware(authService authService, userService userService) mux.MiddlewareFunc {
	return func(next httpPkg.Handler) httpPkg.Handler {
		return httpPkg.HandlerFunc(func(w httpPkg.ResponseWriter, r *httpPkg.Request) {
			// todo check auth

			next.ServeHTTP(w, r)
		})
	}
}
