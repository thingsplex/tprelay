package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/cloud"
	"github.com/thingsplex/tprelay/pkg/cloud/tunnel"
)

func main() {
	log.SetLevel(log.DebugLevel)
	config := cloud.Config{
		BindAddress: ":8083",
	}

	tunMan := tunnel.NewManager()

	server := cloud.NewServer(config,tunMan)
	server.Configure()
	server.StartServer()

}
