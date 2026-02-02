package main

import (
	"os"

	"github.com/stevensopi/Configo/application"
	"github.com/stevensopi/Configo/config"
	"github.com/stevensopi/Configo/logger"
)

func main() {
	config, err := config.LoadConfig()
	logger := logger.NewLogger()

	if err != nil {
		logger.Error("Could not load config", "error", err)
		os.Exit(1)
	}

	app, err := application.NewApplication(&config)
	if err != nil {
		logger.Error("Could not start application", "error", err)
		os.Exit(1)
	}
	app.Run()
}
