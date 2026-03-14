package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 2 * time.Minute
	shutdownTimeout = 15 * time.Second
)

type Server struct {
	*http.Server
}

func NewServer(cfg Config, h http.Handler) *Server {
	return &Server{
		Server: &http.Server{
			Addr:         net.JoinHostPort(cfg.ServerHost, strconv.Itoa(int(cfg.ServerPort))),
			Handler:      h,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		},
	}
}

func (s *Server) Start() error {
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.Addr, err)
	}
	slog.Info("http server started listening", "addr", l.Addr().String())
	go func() {
		if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
		}
	}()
	return nil
}

func (s *Server) Stop() {
	slog.Info("http server shutting down")
	ctx, cancelCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelCtx()
	if err := s.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "err", err)
	}
}
