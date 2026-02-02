package http_server

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
	addr   string
	server *http.Server
	router *mux.Router
	logger *slog.Logger
	raft   *raft_node.RaftNode
}

func NewHttpServer(addr string, logger *slog.Logger, raft *raft_node.RaftNode) *HttpServer {
	router := mux.NewRouter()
	s := &HttpServer{
		addr:   addr,
		router: router,
		raft:   raft,
	}

	s.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	s.logger = logger

	s.registerRaftHandlers()
	s.registerConfigHandlers()

	return s
}

func (s *HttpServer) registerRaftHandlers() {
	s.router.HandleFunc("/raft/add-voter", s.handleAddVoter).Methods("POST")
}

func (s *HttpServer) handleAddVoter(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to add voter", "error", err)
		http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req AddVoterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Error("Failed to add voter", "error", err)
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}

	// Check if this node is leader
	if s.raft.Raft.State() != raft.Leader {
		leader := s.raft.Raft.Leader()
		if leader == "" {
			s.logger.Error("Failed to add voter since no leader is elected")
			http.Error(w, "no leader currently elected", http.StatusServiceUnavailable)
			return
		}

		res := AddVoterResponse{
			Addr: string(leader),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTemporaryRedirect)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			s.logger.Error("Failed to add voter", req.ID, req.Addr)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
		s.logger.Error("Failed to add voter", req.ID, req.Addr)
		http.Error(w, fmt.Sprintf("failed to add voter: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info("Voter added successfully", req.ID, req.Addr)
	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	if s.raft.Raft.State() != raft.Leader {
		s.logger.Error("Follower cannot process PUT request")

		leader := s.raft.Raft.Leader()
		if leader == "" {
			s.logger.Error("Failed to process request since no leader is elected")
			http.Error(w, "no leader currently elected", http.StatusServiceUnavailable)
			return
		}

		res := AddVoterResponse{
			Addr: string(leader),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTemporaryRedirect)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			s.logger.Error("Follower encode response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	id := mux.Vars(r)["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Failed to process PUT request", "error", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	cmd := raft_node.Command{
		Key:   id,
		Value: body,
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		s.logger.Error("Failed to process PUT request", "error", err)
		http.Error(w, "failed to marshal command", http.StatusInternalServerError)
		return
	}

	future := s.raft.Raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		s.logger.Error("Failed to process PUT request", "error", err)
		http.Error(w, fmt.Sprintf("raft apply failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info("PUT request handled successfully")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("config stored"))
}

func (s *HttpServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	val, err := s.raft.FSM.Store.Get(id)
	if err != nil {
		s.logger.Error("Failed to process GET request", "error", err)
		http.Error(w, "config not found", http.StatusNotFound)
		return
	}

	s.logger.Error("GET request handled successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

func (s *HttpServer) handleListConfig(w http.ResponseWriter, r *http.Request) {
	keys, err := s.raft.FSM.Store.List()

	if err != nil {
		s.logger.Error("Failed to process GET request", "error", err)
		http.Error(w, "failed to list configs", http.StatusInternalServerError)
		return
	}

	s.logger.Info("GET request handled successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (s *HttpServer) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	if s.raft.Raft.State() != raft.Leader {
		s.logger.Error("Follower cannot process DELETE request")

		leader := s.raft.Raft.Leader()
		if leader == "" {
			s.logger.Error("Failed to process DELETE request since no leader is elected")

			http.Error(w, "no leader currently elected", http.StatusServiceUnavailable)
			return
		}

		res := AddVoterResponse{
			Addr: string(leader),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTemporaryRedirect)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			s.logger.Error("Follower encode response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	id := mux.Vars(r)["id"]

	cmd := raft_node.Command{
		Key:   id,
		Value: nil, // tombstone
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		s.logger.Error("Failed to process DELETE request", "error", err)
		http.Error(w, "failed to marshal command", http.StatusInternalServerError)
		return
	}

	future := s.raft.Raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		s.logger.Error("Failed to process DELETE request", "error", err)
		http.Error(w, fmt.Sprintf("raft apply failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Error("DELETE request handled successfully")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("config deleted"))

}

func (s *HttpServer) registerConfigHandlers() {
	s.router.HandleFunc("/config/list", s.handleListConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", s.handleGetConfig).Methods("GET")
	s.router.HandleFunc("/config/{id}", s.handlePutConfig).Methods("PUT")
	s.router.HandleFunc("/config/{id}", s.handleDeleteConfig).Methods("DELETE")
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
