package raft_node

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/hashicorp/raft"
	"github.com/stevensopi/Configo/storage"
)

type Command struct {
	Key   string
	Value []byte
}

type FSM struct {
	store  *storage.Store
	logger *slog.Logger
}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	if len(logEntry.Data) == 0 {
		return nil
	}

	f.logger.Info("Received entry on store")

	var cmd Command
	err := json.Unmarshal(logEntry.Data, &cmd)
	if err != nil {
		f.logger.Error("Failed to unmarshal raft command", "error", err)
		return err
	}

	if err := f.store.Store(cmd.Key, cmd.Value); err != nil {
		f.logger.Error("Failed to apply raft command", "error", err)
		return err
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshot{}, nil
}

func (f *FSM) Restore(snapshot io.ReadCloser) error {
	return nil
}

type FSMSnapshot struct{}

func (f *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	sink.Close()
	return nil
}

func (f *FSMSnapshot) Release() {}
