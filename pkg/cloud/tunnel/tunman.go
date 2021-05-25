package tunnel

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type Manager struct {
	tunnels sync.Map
}

func NewManager() *Manager {
	return &Manager{}
}

func (man *Manager) RegisterEdgeConnection(id string, edgeConnection *websocket.Conn,token string,ipAddr string,authConfig AuthConfig) {
	tun := NewWsTunnel(token,ipAddr,edgeConnection,authConfig)
	man.tunnels.Store(id,tun)
	go tun.StartEdgeMsgReader()
}

func (man *Manager) GetTunnelById(id string ) (*WsTunnel,error) {

	connI,ok := man.tunnels.Load(id)
	if !ok {
		return nil,fmt.Errorf("conn not found")
	}

	conn,ok := connI.(*WsTunnel)
	if !ok {
		return nil,fmt.Errorf("invalid connection")
	}
	return conn,nil

}
