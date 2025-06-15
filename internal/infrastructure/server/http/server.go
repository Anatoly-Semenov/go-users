package http

import (
	"context"
	"net/http"
	"time"

	"github.com/anatoly_dev/go-users/internal/app"
	"github.com/anatoly_dev/go-users/internal/config"
	"github.com/anatoly_dev/go-users/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	server  *http.Server
	handler *UserHandler
	config  *config.Config
}

func NewServer(userService *app.UserService, config *config.Config) *Server {
	mux := http.NewServeMux()
	handler := NewUserHandler(userService)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + config.Server.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		server:  server,
		handler: handler,
		config:  config,
	}
}

func (s *Server) Start() error {
	logger.Info("Starting HTTP server", zap.String("port", s.config.Server.HTTPPort))
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}
