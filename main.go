package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stevensopi/Configo/logger"
	"github.com/stevensopi/Configo/server/http"
)

func main() {
	logger := logger.NewLogger()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 1)
	server := http.NewHttpServer(false, ":8080", logger.With("component", "http-server"), nil)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		logger.Error("Server failled", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown failed", "err", err)
	}

	logger.Info("Server stopped")
}
