package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/stevensopi/Configo/server/http_server"
)

type RaftClientConfig struct {
	MaxRetries                int
	RaftSeedMgmtServerAddress string
	Timeout                   time.Duration
}

type RaftClient struct {
	config RaftClientConfig
	client *http.Client
}

func NewRaftClient(config RaftClientConfig) *RaftClient {
	return &RaftClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (rc *RaftClient) AddNodeAsVoter(r *http_server.AddVoterRequest) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// [TODO] need to check response in case new leader is elected and need to try to connect to him
	for i := 0; i < rc.config.MaxRetries; i++ {
		res, err := rc.client.Post(rc.config.RaftSeedMgmtServerAddress, "application/json", bytes.NewBuffer(data))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			return nil
		}

		time.Sleep(time.Second) // sleep for a bit before retrying
	}

	return errors.New("Could not add node as voter to leader")
}
