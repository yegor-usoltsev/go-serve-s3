package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	readTimeout     = 2 * time.Second
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
			Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
			Handler:      h,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
}

func (s *Server) StartAsync() {
	go func() {
		log.Print("Server is starting on http://", s.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Unable to serve: ", err)
		}
	}()
}

func (s *Server) Stop() {
	log.Print("Server is shutting down")
	ctx, cancelCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelCtx()
	if err := s.Shutdown(ctx); err != nil {
		log.Print("Server failed to shut down properly: ", err)
	} else {
		log.Print("Server shut down properly")
	}
}
