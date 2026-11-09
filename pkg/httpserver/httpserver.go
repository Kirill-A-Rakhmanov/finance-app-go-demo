package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Port           string   `envconfig:"HTTP_PORT" default:"3100"`
	TrustedProxies []string `envconfig:"HTTP_TRUSTED_PROXIES" required:"true"`
}

type Server struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

func New(handler http.Handler, c Config) *Server {
	addr := fmt.Sprintf(":%s", c.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 2,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 120,
		MaxHeaderBytes:    1024 * 1024,
	}

	return &Server{
		server:          srv,
		shutdownTimeout: time.Second * 30,
	}
}

func (s *Server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
