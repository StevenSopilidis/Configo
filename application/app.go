package application

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stevensopi/Configo/client"
	"github.com/stevensopi/Configo/config"
	"github.com/stevensopi/Configo/logger"
	"github.com/stevensopi/Configo/raft_node"
	"github.com/stevensopi/Configo/server"
	"github.com/stevensopi/Configo/server/http_server"
)

type Application struct {
	node       *raft_node.RaftNode
	server     server.Server
	raftClient *client.RaftClient
	logger     *slog.Logger
	config     *config.Config
}

func NewApplication(config *config.Config) (*Application, error) {
	logger := logger.NewLogger()

	node, err := raft_node.NewRaftNode(&raft_node.RaftNodeConfig{
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
		logger.Error("could not create raft cluster", "error", err)
		return nil, err
	}

	// if node is first node bootstrap raft cluster
	if config.IsFirstNodeInCluster {
		err := node.BootstrapCluster()
		if err != nil {
			logger.Error("failed to boostrap raft cluster", "error", err)
			return nil, err
		}
	}

	server := http_server.NewHttpServer(":8080", logger.With("component", "http-server"), node)

	client := client.NewRaftClient(client.RaftClientConfig{
		Timeout:                   config.RaftTcpTransportTimeout,
		MaxRetries:                config.RaftMaxConnectionRetries,
		RaftSeedMgmtServerAddress: config.RaftSeedMgmtServerAddress,
	})

	return &Application{
		node:       node,
		server:     server,
		raftClient: client,
		logger:     logger,
		config:     config,
	}, nil
}

func (app *Application) Run() {
	// context for getting notified when we receive sigterm signal
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// start server
	errCh := make(chan error, 1)
	go func() {
		if err := app.server.Start(); err != nil {
			errCh <- err
		}
	}()

	// if not first node in cluster try to connect to cluster
	if !app.config.IsFirstNodeInCluster {
		go func() {
			time.Sleep(time.Second * 10) // sleep for a bit
			err := app.raftClient.AddNodeAsVoter(&http_server.AddVoterRequest{
				ID:   app.config.RaftNodeServerId,
				Addr: app.config.RaftAddr,
			})

			if err != nil {
				errCh <- err
			}

		}()
	}

	select {
	case <-ctx.Done():
		app.logger.Info("Shutdown signal received")
	case err := <-errCh:
		app.logger.Error("Server failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := app.server.Stop(shutdownCtx); err != nil {
		app.logger.Error("Graceful shutdown failed", "err", err)
	}

	app.logger.Info("Server stopped")
}
