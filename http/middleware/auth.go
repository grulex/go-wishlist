package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	authPkg "github.com/grulex/go-wishlist/pkg/auth"
	userPkg "github.com/grulex/go-wishlist/pkg/user"
	wishlistPkg "github.com/grulex/go-wishlist/pkg/wishlist"
	"gopkg.in/guregu/null.v4"
	httpPkg "net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type authService interface {
	Get(ctx context.Context, method authPkg.Method, socialID authPkg.SocialID) (*authPkg.Auth, error)
	Create(ctx context.Context, auth *authPkg.Auth) error
}

type userService interface {
	Create(ctx context.Context, user *userPkg.User) error
}

type wishlistService interface {
	Create(ctx context.Context, wishlist *wishlistPkg.Wishlist) error
}

type telegramUser struct {
	ID              int    `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Username        string `json:"username"`
	LanguageCode    string `json:"language_code"`
	IsPremium       bool   `json:"is_premium"`
	AllowsWriteToPm bool   `json:"allows_write_to_pm"`
}

func NewTelegramAuthMiddleware(
	authService authService,
	userService userService,
	wishlistService wishlistService,
	telegramBotToken string,
) mux.MiddlewareFunc {
	return func(next httpPkg.Handler) httpPkg.Handler {
		return httpPkg.HandlerFunc(func(w httpPkg.ResponseWriter, r *httpPkg.Request) {
			query := r.URL.Query()
			fmt.Println(query)

			hash := query.Get("hash")
			if len(hash) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			authCheckString, err := getAuthCheckString(query)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			secretKey := getHmac256Signature([]byte("WebAppData"), []byte(telegramBotToken))
			expectedHash := getHmac256Signature(secretKey, []byte(authCheckString))
			expectedHashString := hex.EncodeToString(expectedHash)

			if expectedHashString != hash {
				next.ServeHTTP(w, r)
				return
			}

			tgUser := telegramUser{}
			err = json.Unmarshal([]byte(query.Get("user")), &tgUser)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			socialID := authPkg.SocialID(null.NewString(strconv.Itoa(tgUser.ID), true))
			auth, err := authService.Get(r.Context(), authPkg.MethodTelegram, socialID)
			if err != nil && !errors.Is(err, authPkg.ErrNotFound) {
				next.ServeHTTP(w, r)
				return
			}
			if auth == nil {
				auth, err = registerTelegramUser(r.Context(), userService, authService, wishlistService, tgUser)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}
			}

			r = r.WithContext(authPkg.NewContext(r.Context(), auth))
			next.ServeHTTP(w, r)
		})
	}
}

// get alphabetic sorted query string
func getAuthCheckString(values url.Values) (string, error) {
	paramKeys := make([]string, 0)
	for key, v := range values {
		if key == "hash" {
			continue
		}
		if len(v) != 1 {
			return "", errors.New("is not a valid auth query")
		}
		paramKeys = append(paramKeys, key)
	}

	// sort keys
	sort.Strings(paramKeys)

	dataCheckArr := make([]string, len(paramKeys))
	for i, key := range paramKeys {
		dataCheckArr[i] = key + "=" + values.Get(key)
	}

	return strings.Join(dataCheckArr, "\n"), nil
}

func getHmac256Signature(secretKey []byte, data []byte) []byte {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write(data)
	sum := mac.Sum(nil)
	return sum
}

func registerTelegramUser(
	ctx context.Context,
	userService userService,
	authService authService,
	wService wishlistService,
	tgUser telegramUser,
) (*authPkg.Auth, error) {
	user := &userPkg.User{
		FullName: tgUser.FirstName + tgUser.LastName,
	}

	err := userService.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	newWishlist := &wishlistPkg.Wishlist{
		UserID:      user.ID,
		IsDefault:   true,
		Title:       user.FullName + "'s wishlist",
		Description: "",
		IsArchived:  false,
	}

	err = wService.Create(ctx, newWishlist)
	if err != nil {
		return nil, err
	}

	auth := &authPkg.Auth{
		Method:   authPkg.MethodTelegram,
		SocialID: authPkg.SocialID(null.NewString(strconv.Itoa(tgUser.ID), true)),
		UserID:   user.ID,
	}
	err = authService.Create(ctx, auth)
	if err != nil {
		return nil, err
	}

	return auth, nil
}
