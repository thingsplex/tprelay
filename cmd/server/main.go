package main

import (
	"github.com/thingsplex/tprelay/pkg/cloud"
	"github.com/thingsplex/tprelay/pkg/cloud/tunnel"
)

func main() {
	config := cloud.Config{
		BindAddress: ":8083",
	}

	tunMan := tunnel.NewManager()

	server := cloud.NewServer(config,tunMan)
	server.Configure()
	server.StartServer()

}
