package httpserver

import "net/http"

type DatabaseConfigServer struct {
	DatabaseName string
	DatabaseType string
	DatabaseHost string
	DatabasePort int
	Username     string
	Password     string
	OtherDSN     string
	httpServer   http.Server
}
