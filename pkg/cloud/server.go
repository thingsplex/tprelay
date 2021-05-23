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

func (serv *Server) Configure() error {
	var err error
	log.Info("<HttpConn> Configuring HTTP router.")
	serv.server = &http.Server{Addr: serv.config.BindAddress}
	serv.router = mux.NewRouter()
	serv.router.HandleFunc("/edge/{id}/register", serv.edgeRegistrationHandler)
	serv.router.HandleFunc("/cloud/{id}/rest", serv.cloudRestHandler)
	serv.router.HandleFunc("/cloud/{id}/ws", serv.cloudWsHandler)
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
	edgeConnId := vars["id"]
	if edgeConnId == "" {
		return
	}

	token := ""
	ip := ""

	authConf := tunnel.AuthConfig{}

	ws, err := brUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("<HttpConn> Can't upgrade to WS . Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serv.tunMan.RegisterEdgeConnection(edgeConnId,ws,token,ip,authConf)
}

func (serv *Server) cloudRestHandler(w http.ResponseWriter, r *http.Request) {

}

func (serv *Server) cloudWsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> WS connection failed with PANIC")
			log.Error(string(debug.Stack()))
		}
	}()
	vars := mux.Vars(r)
	edgeConnId := vars["id"]

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