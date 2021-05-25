package edge

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"google.golang.org/protobuf/proto"

	//log "github.com/sirupsen/logrus"
)


type TunClient struct {
	address string
	wsConn *websocket.Conn
}

func (tc *TunClient) Connect() error {
	var err error
	tc.wsConn, _, err = websocket.DefaultDialer.Dial(tc.address, nil)
	return err
}

func (tc *TunClient) Send(msg *tunframe.TunnelFrame) error {
	var binMsg []byte
	if err := proto.Unmarshal(binMsg,msg); err != nil {
		log.Info("Failed to parse proto message")
		return err
	}
	return tc.wsConn.WriteMessage(websocket.BinaryMessage,binMsg)
}