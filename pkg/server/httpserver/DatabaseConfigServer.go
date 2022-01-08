package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yulei-gateway/yulei-gateway-controller/pkg/config"
	locallog "github.com/yulei-gateway/yulei-gateway-controller/pkg/log"
	"github.com/yulei-gateway/yulei-gateway-controller/pkg/storage"
)

//HttpConfigServer the http server
//TODO: need rebuild this to use the storage interface and the config model need as a package
type HttpConfigServer struct {
	DatabaseName string
	DatabaseType string
	DatabaseHost string
	DatabasePort int
	Username     string
	Password     string
	OtherDSN     string
	httpServer   *http.Server
	router       *mux.Router
	dataStorage  storage.ControllerStorage
	log          *locallog.LocalLogger
}

func NewHttpConfigServer(conf *config.ServerConfig, log *locallog.LocalLogger, dataStorage storage.ControllerStorage) (*HttpConfigServer, error) {
	if conf == nil || conf.DatabaseConfig == nil {
		return nil, fmt.Errorf("database config can not be nil")
	}
	if dataStorage == nil {
		panic("the data storage can not be nil")
	}
	serviceConfig := &HttpConfigServer{}
	serviceConfig.dataStorage = dataStorage
	var address = conf.BindIP
	if address == "" {
		address = "0.0.0.0"
	}
	address = fmt.Sprintf("%s:%d", address, conf.Port)
	serviceConfig.httpServer = &http.Server{
		Addr: address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	serviceConfig.log = log
	serviceConfig.buildRouter()
	return serviceConfig, nil
}

func (s *HttpConfigServer) buildRouter() {
	s.router = mux.NewRouter()
	//TODO: ADD HTTP HANDLER
	s.httpServer.Handler = s.router
}

func (s *HttpConfigServer) Run(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			s.log.Errorf("recover from error %s", err)
		}
	}()
	log.Fatal(s.httpServer.ListenAndServe())
}

//TODO: ADD CLUSTER LISTEN ROUTE CONFIG
//TODO: ADD DEFAULT PLUGIN SCHEMA
//TODO: ADD PLUGINS ROUTE OR LISTEN BIND
