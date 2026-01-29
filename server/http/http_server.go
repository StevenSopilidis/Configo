package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
	"github.com/stevensopi/Configo/raft_node"
)

type HttpServer struct {
	addr               string
	server             *http.Server
	router             *mux.Router
	logger             *slog.Logger
	isRaftLeaderServer bool // if its raft leader Server it will expose different endpoints
	raft               *raft_node.RaftNode
}

func NewHttpServer(isRaftLeaderServer bool, addr string, logger *slog.Logger, raft *raft_node.RaftNode) *HttpServer {
	router := mux.NewRouter()
	s := &HttpServer{
		addr:   addr,
		router: router,
		raft:   raft,
	}

	s.server = &http.Server{
		Addr: addr,
	}

	s.logger = logger

	s.registerRaftHandlers()
	s.registerConfigHandlers()

	return s
}

func (s *HttpServer) registerRaftHandlers() {
	s.router.HandleFunc("/raft/add-voter", s.handleAddVoter).Methods("POST")
}

type AddVoterRequest struct {
	ID   string `json:"id"`
	Addr string `json:"addr"`
}

func (s *HttpServer) handleAddVoter(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req AddVoterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}

	// Check if this node is leader
	if s.raft.Raft.State() != raft.Leader { // node not leader deny request
		leader := s.raft.Raft.Leader()
		if leader == "" {
			http.Error(w, "no leader currently elected", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, fmt.Sprintf("not leader, current leader: %s", leader), http.StatusTemporaryRedirect)
		return
	}

	// Add voter
	future := s.raft.Raft.AddVoter(
		raft.ServerID(req.ID),
		raft.ServerAddress(req.Addr),
		0,
		10*time.Second,
	)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("failed to add voter: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("voter %s at %s added successfully", req.ID, req.Addr)))
}

func (s *HttpServer) registerConfigHandlers() {
	s.router.HandleFunc("/config/list", handleListConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", handleListConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", handlePutConfig).Methods("PUT")
	s.router.HandleFunc("/config/{id}", handleDeleteConfig).Methods("DELETE")
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
