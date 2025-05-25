package api

import (
	"encoding/json"
	"net/http"
	"carcassonne-ws/internal/websocket"
	"github.com/gorilla/mux"
)

// Server represents the HTTP server
type Server struct {
	hub *websocket.Hub
}

// NewServer creates a new HTTP server
func NewServer(hub *websocket.Hub) *Server {
	return &Server{
		hub: hub,
	}
}

// SetupRoutes sets up the HTTP routes
func (s *Server) SetupRoutes() *mux.Router {
	router := mux.NewRouter()
	
	// Health check endpoint
	router.HandleFunc("/health", s.healthHandler).Methods("GET")
	
	// Room management endpoints (HTTP fallback)
	router.HandleFunc("/api/rooms", s.listRoomsHandler).Methods("GET")
	
	// WebSocket endpoint
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(s.hub, w, r)
	})
	
	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
	
	return router
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status": "healthy",
		"service": "carcassonne-ws",
		"version": "1.0.0",
	}
	
	json.NewEncoder(w).Encode(response)
}

// listRoomsHandler handles room listing requests (HTTP fallback)
func (s *Server) listRoomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// This is a simplified version - in a real implementation,
	// you'd need to access the room manager through the hub
	response := map[string]interface{}{
		"rooms": []interface{}{},
		"message": "Use WebSocket connection for full functionality",
	}
	
	json.NewEncoder(w).Encode(response)
}

// startGameHandler handles game start requests
func (s *Server) startGameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var request struct {
		RoomID   string `json:"roomId"`
		PlayerID string `json:"playerId"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	err := s.hub.StartGame(request.RoomID, request.PlayerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]string{
		"status": "Game started successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}
