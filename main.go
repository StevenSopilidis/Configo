package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stevensopi/Configo/logger"
	"github.com/stevensopi/Configo/raft"
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

	server := http.NewHttpServer(":8080", logger.With("component", "http-server"))

	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	node, err := raft.NewRaftNode(&raft.RaftNodeConfig{
		InternalStorageLocation: "./config-store/",
		RaftLogName:             "store",
		SnapshotLocation:        "./snapshots/",
		NodeServerId:            "node-1",
		SnapshotRetainNum:       2,
		SnapshotLogOutput:       os.Stdout,
		Addr:                    "127.0.0.1:9090",
		TcpTransportPool:        3,
		TcpTransportTimeout:     time.Second * 30,
	}, logger.With("component", "raft-node-1"))

	if err != nil {
		errCh <- err
	} else {
		node.BootstrapCluster()
	}

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
