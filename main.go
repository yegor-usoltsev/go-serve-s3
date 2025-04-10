package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelCtx()

	cfg := NewConfig()
	h := NewHandler(cfg)
	srv := NewServer(cfg, h)
	defer srv.Stop()
	srv.StartAsync()

	<-ctx.Done()
}
