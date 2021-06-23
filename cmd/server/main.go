package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/cloud"
	"github.com/thingsplex/tprelay/pkg/cloud/tunnel"
	"os"
)


var Version string

func main() {
	log.SetLevel(log.DebugLevel)
	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = ":8090"
	}

	config := cloud.Config{
		BindAddress: bindAddress,
	}

	log.Infof("------ Starting tprelay v = %s",Version)

	tunMan := tunnel.NewManager()

	server := cloud.NewServer(config,tunMan)
	server.Configure()
	server.StartServer()

}
