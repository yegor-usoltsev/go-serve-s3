package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic", "error", r)
			os.Exit(1) //nolint:forbidigo
		}
	}()

	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelCtx()

	cfg := NewConfigFromEnv()
	h := NewHandler(cfg)
	srv := NewServer(cfg, h)
	srv.StartAsync()
	defer srv.Stop()

	<-ctx.Done()
}
