package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stevensopi/Configo/logger"
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

	logger := logger.NewLogger()

	for i := 0; i < rc.config.MaxRetries; i++ {
		logger.Info("Sending connection request to", "address", rc.config.RaftSeedMgmtServerAddress)

		res, err := rc.client.Post(rc.config.RaftSeedMgmtServerAddress, "application/json", bytes.NewBuffer(data))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		logger.Info("When trying to become voter received status code: ", "code", res.StatusCode)

		if res.StatusCode == http.StatusOK {
			return nil
		}

		if res.StatusCode == http.StatusTemporaryRedirect {
			// parse received leader addr
			var addVoterRes http_server.AddVoterResponse
			if err := json.NewDecoder(res.Body).Decode(&addVoterRes); err != nil {
				err = fmt.Errorf("Failed to decode AddVoterResponse from leader")
			}

			rc.config.RaftSeedMgmtServerAddress = addVoterRes.Addr
		}

		time.Sleep(time.Second) // sleep for a bit before retrying
	}

	return fmt.Errorf("Failed to add node as voter")
}
