package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"carcassonne-ws/internal/game"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second
	
	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second
	
	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10
	
	// Maximum message size allowed from peer
	maxMessageSize = 512
	
	// Latency ping interval for custom ping/pong
	latencyPingInterval = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should check the origin properly
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	// The websocket connection
	conn *websocket.Conn
	
	// Buffered channel of outbound messages
	send chan []byte
	
	// The hub that manages this client
	hub *Hub
	
	// Player information
	Player *game.Player
	
	// Current room ID
	RoomID string
	
	// Latency tracking
	latency      time.Duration
	lastPingTime time.Time
	latencyMutex sync.RWMutex
	
	// Client ID for ping/pong tracking
	clientID string
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      hub,
		clientID: generateClientID(),
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return "client_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// GetLatency returns the current latency
func (c *Client) GetLatency() time.Duration {
	c.latencyMutex.RLock()
	defer c.latencyMutex.RUnlock()
	return c.latency
}

// updateLatency updates the latency measurement
func (c *Client) updateLatency(pingTimestamp, pongTimestamp int64) {
	c.latencyMutex.Lock()
	defer c.latencyMutex.Unlock()
	
	latency := time.Duration(pongTimestamp - pingTimestamp)
	c.latency = latency
	
	log.Printf("Client %s latency: %v", c.clientID, latency)
}

// sendLatencyPing sends a custom ping message for latency measurement
func (c *Client) sendLatencyPing() {
	pingMsg, err := NewPingMessage(c.clientID)
	if err != nil {
		log.Printf("Error creating ping message: %v", err)
		return
	}
	
	c.lastPingTime = time.Now()
	c.SendMessage(pingMsg)
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		
		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}
		
		// Handle pong messages for latency calculation
		if msg.Type == MessagePong {
			c.handlePongMessage(&msg)
			continue
		}
		
		// Handle the message
		c.hub.handleMessage(c, &msg)
	}
}

// handlePongMessage handles pong messages for latency calculation
func (c *Client) handlePongMessage(msg *Message) {
	var data PongData
	if err := ParseMessage(msg, &data); err != nil {
		log.Printf("Error parsing pong message: %v", err)
		return
	}
	
	// Calculate latency
	c.updateLatency(data.PingTimestamp, data.PongTimestamp)
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	latencyTicker := time.NewTicker(latencyPingInterval)
	defer func() {
		ticker.Stop()
		latencyTicker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			// Standard WebSocket ping for connection keepalive
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			
		case <-latencyTicker.C:
			// Custom ping for latency measurement
			c.sendLatencyPing()
		}
	}
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg *Message) error {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	select {
	case c.send <- messageBytes:
	default:
		close(c.send)
		return err
	}
	
	return nil
}

// SendError sends an error message to the client
func (c *Client) SendError(code, message string) {
	errorMsg, err := NewErrorMessage(code, message)
	if err != nil {
		log.Printf("error creating error message: %v", err)
		return
	}
	
	c.SendMessage(errorMsg)
}

// GetClientID returns the client ID
func (c *Client) GetClientID() string {
	return c.clientID
}

// GetLatencyStats returns latency statistics
func (c *Client) GetLatencyStats() map[string]interface{} {
	c.latencyMutex.RLock()
	defer c.latencyMutex.RUnlock()
	
	return map[string]interface{}{
		"clientId":        c.clientID,
		"latency":         c.latency.String(),
		"latencyMs":       float64(c.latency.Nanoseconds()) / 1e6,
		"lastPingTime":    c.lastPingTime,
		"connectionTime":  time.Since(c.lastPingTime),
	}
}

// Close closes the client connection
func (c *Client) Close() {
	close(c.send)
}

// ServeWS handles websocket requests from the peer
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	
	client := NewClient(hub, conn)
	client.hub.register <- client
	
	log.Printf("New WebSocket connection established: %s", client.clientID)
	
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines
	go client.writePump()
	go client.readPump()
}
