package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/grulex/go-wishlist/config"
	"github.com/grulex/go-wishlist/container"
	httpUtil "github.com/grulex/go-wishlist/http/httputil"
	"github.com/grulex/go-wishlist/http/middleware"
	"github.com/grulex/go-wishlist/http/usecase"
	"github.com/rs/cors"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(listenAddr string, container *container.ServiceContainer, config *config.Config) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/health", httpUtil.ResponseWrapper(usecase.MakeUseCaseHealthCheck())).Methods("GET")
	//r.HandleFunc("/index", httpUtil.ResponseWrapper(usecase.MakeUseCaseIndex())).Methods("GET")

	authMiddleware := middleware.NewTelegramAuthMiddleware(container.Auth, container.User, container.Wishlist, config.TelegramBotToken)
	r.Use(authMiddleware)

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
