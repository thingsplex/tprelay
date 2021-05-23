package tunnel

import (
	"github.com/gorilla/websocket"
	"github.com/thingsplex/tprelay/pkg/utils"
	"net/http"
	"sync"
	"time"
)

type Bridge struct {
	State                uint8 // new -> edge -> cloud -> new
	StartedAt            time.Time
	ExpiresAt            time.Time
	Token                string // Rendezvous token used for matching Cloud connection with Edge connection
	EdgeConnIp           string
	EdgeConnection       *websocket.Conn // only one connection per bridge
	liveSyncTransactions sync.Map //
	cloudWsConnections   sync.Map // Multiple
	CloudConnAuth        AuthConfig // Must be set by Edge connection
}

type AuthConfig struct {
	AuthMethod          string `json:"omitempty"` // none , bearer , basic ,header-token, query-token
	AuthToken           string `json:"omitempty"` // Bearer token
	AuthUsername        string `json:"omitempty"` // Username for Basic auth
	AuthPassword        string `json:"omitempty"` // Password for Basic auth
	AuthCustomParamName string `json:"omitempty"` // Name of custom header that stores token. Or name of query parameter that holds token.
}

type syncTransaction struct {
	respWriter     http.ResponseWriter // channel is used by HTTP/WS action node for sending response to http request
	startTime      time.Time
	responseSignal chan bool
}

func NewBridge(token string, edgeConnIp string, edgeConnection *websocket.Conn, cloudConnAuth AuthConfig) *Bridge {
	return &Bridge{StartedAt: time.Now(), Token: token, EdgeConnIp: edgeConnIp, EdgeConnection: edgeConnection, CloudConnAuth: cloudConnAuth}
}

func (conn *Bridge) RegisterCloudWsConn(cConn *websocket.Conn) int32  {
	connId := utils.GenerateRandomNumber()
	conn.cloudWsConnections.Store(connId,cConn)
	return connId
}

func (conn *Bridge) RegisterCloudRestConn()  {

}

