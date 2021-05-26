package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/edge"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	log.SetLevel(log.DebugLevel)

	tc := edge.NewTunClient("ws://localhost:8083","edge-dev-1",1,http.Header{})

	if err := tc.Connect(); err != nil {
		log.Errorf("Connect() error = %v, ", err)
	}

	msg := tunframe.TunnelFrame{
		MsgType:   tunframe.TunnelFrame_HTTP_RESP,
		Headers:   nil,
		Params:    nil,
		ReqId:     0,
		CorrId:    0,
		ReqUrl:    "",
		ReqMethod: "",
		RespCode:  0,
		Payload:   []byte("pong"),
	}

	stream := tc.GetStream()

	for {
		newMsg :=<- stream
		log.Info("New msg from stream")
		log.Info("%+v",newMsg)
		msg.CorrId = newMsg.ReqId
		msg.RespCode = 200
		if err := tc.Send(&msg);err != nil {
			log.Error("Failed to send message . Err:",err.Error())
		}
		//if string(newMsg.Payload) == "ping" {
		//
		//}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	_ = <-c

	tc.Close()

}
