package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/config"
	"github.com/grulex/go-wishlist/container"
	httpUtil "github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/middleware"
	"github.com/grulex/go-wishlist/http/usecase"
	"github.com/grulex/go-wishlist/http/usecase/users"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/add_product_to_wishlist"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/book_wishlist_item"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/get_wishlist"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/get_wishlist_items"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/remove_product_from_wishlist"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/subscribe_wishlist"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/unbook_wishlist_item"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/unsubscribe_wishlist"
	"github.com/grulex/go-wishlist/http/usecase/wishlists/update_wishlist"
	"github.com/rs/cors"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(listenAddr string, container *container.ServiceContainer, config *config.Config) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/health", httpUtil.ResponseWrapper(usecase.MakeHealthCheckUsecase())).Methods("GET")
	//r.HandleFunc("/index", httpUtil.ResponseWrapper(usecase.MakeIndexUsecase())).Methods("GET")

	apiRouter := r.PathPrefix("/api").Subrouter()
	authMiddleware := middleware.NewTelegramAuthMiddleware(container.Auth, container.User, container.Wishlist, config.TelegramBotToken)
	apiRouter.Use(authMiddleware)

	apiRouter.HandleFunc("/profile", httpUtil.ResponseWrapper(
		users.MakeGetProfileUsecase(container.Subscribe, container.Wishlist),
	)).Methods("GET")

	apiRouter.HandleFunc("/wishlists/{id}", httpUtil.ResponseWrapper(
		get_wishlist.MakeGetWishlistUsecase(container.Subscribe, container.Wishlist),
	)).Methods("GET")

	apiRouter.HandleFunc("/wishlists/{id}", httpUtil.ResponseWrapper(
		update_wishlist.MakeUpdateWishlistUsecase(container.Wishlist),
	)).Methods("PUT")

	apiRouter.HandleFunc("/wishlists/{id}/subscribe", httpUtil.ResponseWrapper(
		subscribe_wishlist.MakeSubscribeWishlistUsecase(container.Wishlist, container.Subscribe),
	)).Methods("POST")

	apiRouter.HandleFunc("/wishlists/{id}/unsubscribe", httpUtil.ResponseWrapper(
		unsubscribe_wishlist.MakeUnSubscribeWishlistUsecase(container.Wishlist, container.Subscribe),
	)).Methods("POST")

	apiRouter.HandleFunc("/wishlists/{id}/items", httpUtil.ResponseWrapper(
		get_wishlist_items.MakeGetWishlistItemsUsecase(container.Wishlist, container.Product),
	)).Methods("GET")

	apiRouter.HandleFunc("/wishlists/{id}/items", httpUtil.ResponseWrapper(
		add_product_to_wishlist.MakeAddProductToWishlistUsecase(container.Wishlist, container.Product),
	)).Methods("POST")

	apiRouter.HandleFunc("/wishlists/{id}/items/{productId}/book", httpUtil.ResponseWrapper(
		book_wishlist_item.MakeBookWishlistItemUsecase(container.Wishlist),
	)).Methods("PUT")

	apiRouter.HandleFunc("/wishlists/{id}/items/{productId}/book", httpUtil.ResponseWrapper(
		unbook_wishlist_item.MakeUnBookWishlistItemUsecase(container.Wishlist),
	)).Methods("DELETE")

	apiRouter.HandleFunc("/wishlists/{id}/items/{productId}", httpUtil.ResponseWrapper(
		remove_product_from_wishlist.MakeRemoveProductFromWishlistUsecase(container.Wishlist),
	)).Methods("DELETE")

	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	handler := c.Handler(r)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}
	return &Server{
		httpServer: server,
	}
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
