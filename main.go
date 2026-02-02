package main

import (
	"os"
	"time"

	"github.com/stevensopi/Configo/application"
	"github.com/stevensopi/Configo/config"
)

func main() {
	config := config.Config{
		RaftInternalStorageLocation: "./config-store/",
		RaftNodeServerId:            "node-1",
		RaftLogName:                 "store",
		RaftSnapshotLocation:        "./snapshots/",
		RaftSnapshotRetainNum:       2,
		RaftAddr:                    "127.0.0.1:9090",
		RaftTcpTransportPool:        3,
		RaftTcpTransportTimeout:     time.Second * 30,
		RaftMaxConnectionRetries:    5,
		RaftSeedMgmtServerAddress:   "127.0.0.1:9090",
		AppAddr:                     "127.0.0.1:8080",
		IsFirstNodeInCluster:        true,
	}

	app, err := application.NewApplication(&config)
	if err != nil {
		os.Exit(1)
	}

	app.Run()
}
