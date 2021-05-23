package tunnel

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Manager struct {
	tunnels sync.Map
}

func (man *Manager) RegisterEdgeConnection(id string, edgeConnection *websocket.Conn,token string,ipAddr string,authConfig AuthConfig) {
	bridge := NewBridge(token,ipAddr,edgeConnection,authConfig)
	man.tunnels.Store(id,bridge)
}

func (man *Manager) GetTunnelById(id string ) (*Bridge,error) {

	connI,ok := man.tunnels.Load(id)
	if !ok {
		return nil,fmt.Errorf("conn not found")
	}

	conn,ok := connI.(*Bridge)
	if !ok {
		return nil,fmt.Errorf("invalid connection")
	}
	return conn,nil

}
