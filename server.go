package main

import (
	"context"
	"errors"
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
		Server: &http.Server{ //nolint:exhaustruct
			Addr:         net.JoinHostPort(cfg.ServerHost, strconv.Itoa(int(cfg.ServerPort))),
			Handler:      h,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		},
	}
}

func (s *Server) StartAsync() {
	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		l, err := net.Listen("tcp", s.Addr)
		if err != nil {
			slog.Error("http server unable to listen on address", "addr", s.Addr, "error", err)
			errCh <- err
			return
		}
		errCh <- nil
		slog.Info("http server started listening", "addr", l.Addr().String())
		if err := s.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server unable to handle requests", "error", err)
		}
	}()

	if err := <-errCh; err != nil {
		panic(err)
	}
}

func (s *Server) Stop() {
	slog.Info("http server initiating shutdown")
	ctx, cancelCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelCtx()
	if err := s.Shutdown(ctx); err != nil {
		panic(err)
	}
	slog.Info("http server completed shutdown")
}
