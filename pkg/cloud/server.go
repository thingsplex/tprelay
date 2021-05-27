package cloud

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/cloud/tunnel"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	brUpgrader = websocket.Upgrader{
		Subprotocols:     []string{},
		HandshakeTimeout: time.Second * 20,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Server struct {
	server         *http.Server
	router         *mux.Router
	config         Config
	tunMan         *tunnel.Manager
}

func NewServer(config Config, tunMan *tunnel.Manager) *Server {
	return &Server{config: config, tunMan: tunMan}
}

func (serv *Server) Configure() error {
	var err error
	log.Info("<HttpConn> Configuring HTTP router.")
	serv.server = &http.Server{Addr: serv.config.BindAddress}
	serv.router = mux.NewRouter()
	serv.router.HandleFunc("/edge/{tunId}/register", serv.edgeRegistrationHandler) // WS for connection from edge devices.
	serv.router.HandleFunc("/cloud/{tunId}/health", serv.cloudHttpHandler) // Endpoint for cloud HTTP requests.
	serv.router.HandleFunc("/cloud/{tunId}/flow/{flowId}/rest", serv.cloudHttpHandler) // Endpoint for cloud HTTP requests.
	serv.router.HandleFunc("/cloud/{tunId}/flow/{flowId}/ws", serv.cloudWsHandler)     // Endpoint for cloud WS connections.
	serv.router.HandleFunc("/cloud/{tunId}/api/registry/{subComp}", serv.cloudHttpHandler)     // Endpoint for cloud WS connections.
	serv.router.HandleFunc("/cloud/{tunId}/api/flow/context/{flowId}", serv.cloudHttpHandler)     // Endpoint for cloud WS connections.

	serv.server.Handler = serv.router
	log.Info("<HttpConn> HTTP router configured ")
	return err
}

func (serv *Server) StartServer() {
	log.Info("<HttpConn> Starting HTTP server.")
	serv.server.ListenAndServe()
}

func (serv *Server) edgeRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> WS connection failed with PANIC")
			log.Error(string(debug.Stack()))
		}
	}()
	vars := mux.Vars(r)
	edgeConnId := vars["tunId"]
	if edgeConnId == "" {
		return
	}

	edgeToken := GetEdgeToken(r)
	ValidateEdgeToken(edgeToken,w)

	authConf := tunnel.AuthConfig{AuthToken: edgeToken}

	ws, err := brUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("<HttpConn> Can't upgrade to WS . Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serv.tunMan.RegisterEdgeConnection(edgeConnId,ws,edgeToken,"",authConf)
}

func (serv *Server) cloudHttpHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> HTTP request failed with PANIC")
			log.Error(string(debug.Stack()))
		}
	}()
	vars := mux.Vars(r)
	edgeConnId := vars["tunId"]
	if edgeConnId == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	edgeToken := GetEdgeToken(r)
	ValidateEdgeToken(edgeToken,w)

	tunn,err := serv.tunMan.GetTunnelById(edgeConnId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	err = tunn.SendHttpRequestAndWaitForResponse(w,r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write( []byte(err.Error()))
	}

}

func (serv *Server) cloudWsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> WS connection failed with PANIC")
			log.Error(string(debug.Stack()))
		}
	}()
	vars := mux.Vars(r)
	edgeConnId := vars["tunId"]

	if edgeConnId == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	edgeToken := GetEdgeToken(r)
	ValidateEdgeToken(edgeToken,w)

	edgeConn, err := serv.tunMan.GetTunnelById(edgeConnId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ws, err := brUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("<HttpConn> Can't upgrade to WS . Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	edgeConn.RegisterCloudWsConn(ws)
}

func GetEdgeToken(r *http.Request)string {
	uq := r.URL.Query()
	edgeToken := uq.Get("tplex_token")
	if edgeToken == "" {
		edgeToken = r.Header.Get("X-TPlex-Token")
	}
	return edgeToken
}

func ValidateEdgeToken(edgeToken string , w http.ResponseWriter) {
	if edgeToken == ""{
		log.Info("<HttpConn> Edge dev registration for dev rejected , missing tplex token.")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("token not found"))
		return
	}else {
		log.Debug("<HttpConn> Registering new edge connection. ")
	}
}