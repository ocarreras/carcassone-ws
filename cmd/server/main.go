package main

import (
	"log"
	"net/http"
	"os"
	"carcassonne-ws/internal/api"
	"carcassonne-ws/internal/websocket"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create HTTP server
	server := api.NewServer(hub)
	router := server.SetupRoutes()

	log.Printf("Carcassonne WebSocket server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	
	// Start the server
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
