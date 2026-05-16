package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/koblas/besops/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a := app.New(cfg)
	if err := a.Start(ctx); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}
