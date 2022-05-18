package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/cloud"
	"github.com/thingsplex/tprelay/pkg/cloud/tunnel"
	"os"
)

var Version string

func main() {
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
	bindAddress := os.Getenv("BIND_ADDRESS")

	if bindAddress == "" {
		bindAddress = ":8090"
	}

	config := cloud.Config{
		BindAddress: bindAddress,
	}

	log.Infof("------ Starting tprelay v = %s , address = %s", Version, bindAddress)

	tunMan := tunnel.NewManager()

	server := cloud.NewServer(config, tunMan)
	server.Configure()
	server.StartServer()

}
