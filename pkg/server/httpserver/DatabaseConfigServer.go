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
	"github.com/yulei-gateway/yulei-gateway-controller/pkg/storage/driver/db"
)

type HttpConfigServer struct {
	DatabaseName    string
	DatabaseType    string
	DatabaseHost    string
	DatabasePort    int
	Username        string
	Password        string
	OtherDSN        string
	httpServer      *http.Server
	router          *mux.Router
	databaseStorage *db.DatabaseStorage
	log             *locallog.LocalLogger
}

func NewHttpConfigServer(conf *config.ServerConfig, log *locallog.LocalLogger) (*HttpConfigServer, error) {
	if conf == nil || conf.DatabaseConfig == nil {
		return nil, fmt.Errorf("database config can not be nil")
	}
	serviceConfig := &HttpConfigServer{}
	storageConfig := db.NewDatabaseStorage(
		conf.DatabaseConfig.Type,
		conf.DatabaseConfig.Host,
		conf.DatabaseConfig.UserName,
		conf.DatabaseConfig.Password,
		conf.DatabaseConfig.Name,
		conf.DatabaseConfig.Port)
	serviceConfig.databaseStorage = storageConfig
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
