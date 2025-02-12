package main

import (
	"context"
	"log"

	"github.com/dnsoftware/mpm-save-get-shares/config"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/dnsoftware/mpm-shares-processor/app"
	"github.com/dnsoftware/mpm-shares-processor/constants"
)

func main() {
	ctx := context.Background()

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env"

	cfg, err := config.New(configFile, envFile)
	if err != nil {
		log.Fatalf("Main config failed: %s", err.Error())
	}

	filePath, err := logger.GetLoggerMainLogPath()
	if err != nil {
		panic("Bad logger init: " + err.Error())
	}
	logger.InitLogger(logger.LogLevelDebug, filePath)

	err = app.Run(ctx, cfg)

}
