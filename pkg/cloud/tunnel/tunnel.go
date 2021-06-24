package tunnel

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"github.com/thingsplex/tprelay/pkg/utils"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type WsTunnel struct {
	State                uint8 // new -> edge -> cloud -> new
	StartedAt            time.Time
	ExpiresAt            time.Time
	Token                string // Rendezvous token used for matching Cloud connection with Edge connection
	EdgeConnIp           string
	edgeConnection       *websocket.Conn // only one connection per bridge
	liveSyncTransactions sync.Map        //
	cloudWsConnections   sync.Map        // Multiple
	CloudConnAuth        AuthConfig      // Must be set by Edge connection
	IsEdgeConnectionActive bool
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
	responseMsg    *tunframe.TunnelFrame
}

func NewWsTunnel(token string, edgeConnIp string, edgeConnection *websocket.Conn, cloudConnAuth AuthConfig) *WsTunnel {
	return &WsTunnel{StartedAt: time.Now(), Token: token, EdgeConnIp: edgeConnIp, edgeConnection: edgeConnection, CloudConnAuth: cloudConnAuth}
}

func (conn *WsTunnel) RegisterCloudWsConn(cConn *websocket.Conn) int64  {
	connId := utils.GenerateRandomNumber()
	conn.cloudWsConnections.Store(connId,cConn)
	return connId
}

func (conn *WsTunnel) RegisterCloudRestConn()  {

}

// StartEdgeMsgReader start loop that is reading messages from edge connection. It can be only one reader per tunnel.Each inbound message is broadcasted to all
// cloud connections
func (conn *WsTunnel) StartEdgeMsgReader() {
	log.Debug("<edgeConn> Starting edge msg reader")
	for {
		msgType, msg, err := conn.edgeConnection.ReadMessage() // reading message from Edge devices
		if err != nil {
			log.Info("<edgeConn> WS Read error :", err)
			break
		}
		if msgType == websocket.TextMessage {
			log.Info("<edgeConn> TextMessage type is not supported")
		} else if msgType == websocket.BinaryMessage {
			log.Debug("<edgeConn> New binary ws message from tunnel.len=",len(msg))
			tunMsg := tunframe.TunnelFrame{}

			if err := proto.Unmarshal(msg,&tunMsg); err != nil {
				log.Info("Failed to parse proto message")
				continue
			}
			switch tunMsg.MsgType {
			case tunframe.TunnelFrame_HTTP_RESP:
				if tunMsg.CorrId == 0 {
					log.Debug("response message has empty correlation id")
					continue
				}
				tranI,ok := conn.liveSyncTransactions.Load(tunMsg.CorrId)
				if !ok {
					log.Debug("unregistered transaction , id = ",tunMsg.CorrId)
					continue
				}
				trans , ok := tranI.(*syncTransaction)
				if !ok {
					log.Error("Can't cast transaction ")
					continue
				}
				trans.responseMsg = &tunMsg
				trans.responseSignal <- true
				log.Debug("Response signal was sent")

			case tunframe.TunnelFrame_WS_MSG:
				log.Debug("Msg to broadcast :", string(tunMsg.Payload))

				conn.cloudWsConnections.Range(func(key, value interface{}) bool {
					wsConn,ok := value.(*websocket.Conn)
					if !ok {
						return false
					}
					wsConn.WriteMessage(websocket.TextMessage,tunMsg.Payload)
					return true
				})

			default:
				log.Info("Unsupported frame type")

			}
			// 1. unmarshal protobuf encoded msg into struct
			// 2. check message type , either directly route to all connected Cloud connections or store response into sync transactions map and unlock block
		} else {
			log.Debug("<edgeConn> Message of type = ", msgType)
		}
	}
	conn.IsEdgeConnectionActive = false
}

// StartCloudWsMsgReader start loop that is reading message from cloud connection. One loop per connection.
func (conn *WsTunnel) StartCloudWsMsgReader() {
	// TODO : Set limit of allowed count of connection per tunnel.
}

// SendHttpRequestAndWaitForResponse sends http request over tunnel , blocks until it receives response from tunnel or times out
func (conn *WsTunnel) SendHttpRequestAndWaitForResponse(w http.ResponseWriter, r *http.Request) error {
	// TODO : Set limit of allowed count of connection per tunnel.

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	headers := map[string]*tunframe.TunnelFrame_StringArray{}

	for k := range r.Header {
		//TODO : add header filter
		headers[k] = &tunframe.TunnelFrame_StringArray{Items: r.Header.Values(k)}
	}

	reqId := utils.GenerateRandomNumber()

	newMsg := tunframe.TunnelFrame{
		MsgType:   tunframe.TunnelFrame_HTTP_REQ,
		Headers:   headers,
		Params:    nil,
		Vars:      mux.Vars(r),
		ReqId:     reqId,
		CorrId:    0,
		ReqUrl:    r.RequestURI,
		ReqMethod: r.Method,
		Payload:   body,
	}

	binMsg , err := proto.Marshal(&newMsg)
	if err != nil {
		return err
	}

	respSignalCh := make(chan bool)

	syncTransaction := syncTransaction{
		respWriter:     w,
		startTime:      time.Now(),
		responseSignal: respSignalCh,
	}

	conn.liveSyncTransactions.Store(reqId,&syncTransaction)

	// TODO : configure write deadline
	conn.edgeConnection.WriteMessage(websocket.BinaryMessage,binMsg)

	select {
	case <-respSignalCh:

	case <-time.After(15 * time.Second):
		return fmt.Errorf("response timeout")
	}
	if syncTransaction.responseMsg == nil {
		return fmt.Errorf("empty response message")

	}
	code := syncTransaction.responseMsg.RespCode

	if syncTransaction.responseMsg != nil {
		//log.Debug("%+v",syncTransaction.responseMsg)
		if syncTransaction.responseMsg.Headers != nil {
			for k,v := range syncTransaction.responseMsg.Headers {
				it := v.Items
				if len(it)>0 {
					w.Header().Set(k,it[0])
					log.Info("Sending header ",k)
				}
			}
		}
	}
	if code != 200 && code != 0 {
		w.WriteHeader(int(code))
	}

	w.Write(syncTransaction.responseMsg.Payload)

	return nil
}

