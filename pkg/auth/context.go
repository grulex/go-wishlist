package auth

import (
	"context"
)

const contextAuthKey = "auth"

func NewContext(ctx context.Context, auth Auth) context.Context {
	return context.WithValue(ctx, contextAuthKey, auth)
}

func FromContext(ctx context.Context) (Auth, bool) {
	auth, ok := ctx.Value(contextAuthKey).(Auth)
	return auth, ok
}
