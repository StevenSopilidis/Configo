package http

import (
	"context"
	"net/http"

	"log/slog"

	"github.com/gorilla/mux"
)

type HttpServer struct {
	addr   string
	server *http.Server
	router *mux.Router
	logger *slog.Logger
}

func NewHttpServer(addr string, logger *slog.Logger) *HttpServer {
	router := mux.NewRouter()
	s := &HttpServer{
		addr:   addr,
		router: router,
	}

	s.server = &http.Server{
		Addr: addr,
	}

	s.logger = logger

	s.registerHandlers()
	return s
}

func (s *HttpServer) registerHandlers() {
	s.router.HandleFunc("/config/list", HandleListConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", HandleListConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", HandlePutConfig).Methods("PUT")
	s.router.HandleFunc("/config/{id}", HandleDeleteConfig).Methods("DELETE")
}

func (s *HttpServer) Start() error {
	s.logger.Info("Server staring at address", "addr", s.addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *HttpServer) Stop(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	return s.server.Shutdown(ctx)
}
