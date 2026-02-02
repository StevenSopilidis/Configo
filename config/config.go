package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	RaftInternalStorageLocation string `mapstructure:"RAFT_INTERNAL_STORAGE_LOCATION"`
	RaftNodeServerId            string `mapstructure:"RAFT_NODE_SERVER_ID"`
	RaftLogName                 string `mapstructure:"RAFT_LOG_NAME"`
	RaftSnapshotLocation        string `mapstructure:"RAFT_SNAPSHOT_LOCATION"`
	RaftSnapshotRetainNum       int    `mapstructure:"RAFT_SNAPSHOT_RETAIN_NUM"`
	RaftAddr                    string `mapstructure:"RAFT_ADDR"`
	RaftTcpTransportPool        int    `mapstructure:"RAFT_TCP_TRANSPORT_POOL"`
	RaftTcpTransportTimeout     string `mapstructure:"RAFT_TCP_TRANSPORT_TIMEOUT"`
	RaftMaxConnectionRetries    int    `mapstructure:"RAFT_MAX_CONNECTION_RETRIES"`
	RaftSeedMgmtServerAddress   string `mapstructure:"RAFT_SEED_MGMT_SERVER_ADDRESS"`
	AppAddr                     string `mapstructure:"APP_ADDR"`
	IsFirstNodeInCluster        bool   `mapstructure:"IS_FIRST_NODE_IN_CLUSTER"`
}

func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	// Bind each environment variable explicitly
	viper.BindEnv("RAFT_INTERNAL_STORAGE_LOCATION")
	viper.BindEnv("RAFT_NODE_SERVER_ID")
	viper.BindEnv("RAFT_LOG_NAME")
	viper.BindEnv("RAFT_SNAPSHOT_LOCATION")
	viper.BindEnv("RAFT_SNAPSHOT_RETAIN_NUM")
	viper.BindEnv("RAFT_ADDR")
	viper.BindEnv("RAFT_TCP_TRANSPORT_POOL")
	viper.BindEnv("RAFT_TCP_TRANSPORT_TIMEOUT")
	viper.BindEnv("RAFT_SEED_MGMT_SERVER_ADDRESS")
	viper.BindEnv("RAFT_MAX_CONNECTION_RETRIES")
	viper.BindEnv("APP_ADDR")
	viper.BindEnv("LEADER_MGMT_SERVER_ADDRESS")
	viper.BindEnv("IS_FIRST_NODE_IN_CLUSTER")

	err = viper.Unmarshal(&config)
	return config, err
}
