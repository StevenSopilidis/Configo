package raft_node

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/stevensopi/Configo/storage"
)

type RaftNodeConfig struct {
	InternalStorageLocation string
	NodeServerId            string
	RaftLogName             string
	SnapshotLocation        string
	SnapshotRetainNum       int
	SnapshotLogOutput       io.Writer
	Addr                    string
	TcpTransportPool        int
	TcpTransportTimeout     time.Duration
}

type RaftNode struct {
	Raft   *raft.Raft
	FSM    *FSM
	logger *slog.Logger
	id     raft.ServerID
	addr   raft.ServerAddress
}

func NewRaftNode(clusterConfig *RaftNodeConfig, logger *slog.Logger) (*RaftNode, error) {
	store, err := storage.NewStore(clusterConfig.InternalStorageLocation, logger.With("component", "config-store"))
	if err != nil {
		return nil, err
	}

	if clusterConfig.SnapshotRetainNum < 1 {
		logger.Error("Invalid snapshot number provided")
		return nil, errors.New("Snapshot contain number must be atleast one")
	}

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(clusterConfig.NodeServerId)

	boltStore, err := raftboltdb.NewBoltStore(clusterConfig.RaftLogName)
	if err != nil {
		logger.Error("Could not create log store", "error", err)
		return nil, err
	}

	snapStore, err := raft.NewFileSnapshotStore(clusterConfig.SnapshotLocation, clusterConfig.SnapshotRetainNum, clusterConfig.SnapshotLogOutput)
	if err != nil {
		logger.Error("Could not create snapshot store", "error", err)
		return nil, err
	}

	resolved, _ := net.ResolveTCPAddr("tcp", clusterConfig.Addr)
	transport, err := raft.NewTCPTransport(resolved.String(), resolved, clusterConfig.TcpTransportPool, clusterConfig.TcpTransportTimeout, os.Stdout)
	if err != nil {
		logger.Error("Could not create tcp transport", "error", err)
		return nil, err
	}

	fsm := &FSM{
		Store:  store,
		logger: logger.With("component", "fsm"),
	}

	r, err := raft.NewRaft(
		config,
		fsm,
		boltStore, boltStore, snapStore, transport)

	if err != nil {
		logger.Error("Could not create raft", "error", err)
		return nil, err
	}

	return &RaftNode{
		Raft:   r,
		FSM:    fsm,
		logger: logger,
		id:     config.LocalID,
		addr:   transport.LocalAddr(),
	}, nil
}

// should only be called once by the first node joining the cluster
func (node *RaftNode) BootstrapCluster() error {
	node.logger.Info("Bootstraping raft cluster")
	config := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(node.id),
				Address: node.addr,
			},
		},
	}
	future := node.Raft.BootstrapCluster(config)

	return future.Error()
}
