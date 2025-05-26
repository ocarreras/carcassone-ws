package websocket

import (
	"carcassonne-ws/internal/game"
	"carcassonne-ws/internal/room"
	"log"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool
	
	// Inbound messages from the clients
	broadcast chan []byte
	
	// Register requests from the clients
	register chan *Client
	
	// Unregister requests from clients
	unregister chan *Client
	
	// Room manager
	roomManager *room.Manager
	
	// Bot processing ticker
	botTicker *time.Ticker
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		roomManager: room.NewManager(),
		botTicker:   time.NewTicker(2 * time.Second), // Process bot moves every 2 seconds
	}
}

// Run starts the hub
func (h *Hub) Run() {
	go h.processBotMoves()
	
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))
			
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				
				// Handle player leaving
				if client.Player != nil && client.RoomID != "" {
					h.roomManager.LeaveRoom(client.RoomID, client.Player.ID)
					h.broadcastRoomState(client.RoomID)
				}
				
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}
			
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// handleMessage handles incoming messages from clients
func (h *Hub) handleMessage(client *Client, msg *Message) {
	log.Printf("Handling : %s", msg.Type)
	switch msg.Type {
	case MessageConnect:
		h.handleConnect(client, msg)
	case MessageListRooms:
		h.handleListRooms(client, msg)
	case MessageCreateRoom:
		h.handleCreateRoom(client, msg)
	case MessageJoinRoom:
		h.handleJoinRoom(client, msg)
	case MessageLeaveRoom:
		h.handleLeaveRoom(client, msg)
	case MessageAddBot:
		h.handleAddBot(client, msg)
	case MessagePlaceTile:
		h.handlePlaceTile(client, msg)
	case MessagePlaceMeeple:
		h.handlePlaceMeeple(client, msg)
	case MessagePing:
		h.handlePing(client, msg)
	default:
		log.Printf("Invalid message  : %s", msg.Type)
		client.SendError("UNKNOWN_MESSAGE", "Unknown message type")
	}
}

// handleConnect handles player connection
func (h *Hub) handleConnect(client *Client, msg *Message) {
	var data ConnectData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid connect data")
		return
	}
	
	// Create player
	player := &game.Player{
		ID:      data.PlayerID,
		Name:    data.Name,
		Color:   data.Color,
		Meeples: 7,
		IsBot:   false,
		Score:   0,
	}
	
	client.Player = player
	
	// Send room list
	h.handleListRooms(client, msg)
}

// handleListRooms handles room listing request
func (h *Hub) handleListRooms(client *Client, msg *Message) {
	roomInfos := h.roomManager.GetActiveRooms()
	
	// Convert room.RoomInfo to websocket.RoomInfo
	rooms := make([]RoomInfo, len(roomInfos))
	for i, roomInfo := range roomInfos {
		rooms[i] = RoomInfo{
			ID:          roomInfo.ID,
			Name:        roomInfo.Name,
			PlayerCount: roomInfo.PlayerCount,
			MaxPlayers:  roomInfo.MaxPlayers,
			GameStarted: roomInfo.GameStarted,
			CreatedBy:   roomInfo.CreatedBy,
		}
	}
	
	response, err := CreateMessage(MessageListRooms, ListRoomsData{
		Rooms: rooms,
	})
	if err != nil {
		client.SendError("INTERNAL_ERROR", "Failed to create room list")
		return
	}
	
	client.SendMessage(response)
}

// handleCreateRoom handles room creation
func (h *Hub) handleCreateRoom(client *Client, msg *Message) {
	if client.Player == nil {
		client.SendError("NOT_CONNECTED", "Must connect first")
		return
	}
	
	var data CreateRoomData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid create room data")
		return
	}
	
	room, err := h.roomManager.CreateRoom(data.RoomName, client.Player.ID, data.MaxPlayers)
	if err != nil {
		client.SendError("CREATE_FAILED", err.Error())
		return
	}
	
	// Add creator to room
	err = room.AddPlayer(client.Player)
	if err != nil {
		client.SendError("JOIN_FAILED", err.Error())
		return
	}
	
	client.RoomID = room.ID
	
	// Send room state
	h.sendRoomState(client, room)
}

// handleJoinRoom handles joining a room
func (h *Hub) handleJoinRoom(client *Client, msg *Message) {
	if client.Player == nil {
		client.SendError("NOT_CONNECTED", "Must connect first")
		return
	}
	
	var data JoinRoomData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid join room data")
		return
	}
	
	err := h.roomManager.JoinRoom(data.RoomID, client.Player)
	if err != nil {
		client.SendError("JOIN_FAILED", err.Error())
		return
	}
	
	client.RoomID = data.RoomID
	
	// Broadcast room state to all players in room
	h.broadcastRoomState(data.RoomID)
}

// handleLeaveRoom handles leaving a room
func (h *Hub) handleLeaveRoom(client *Client, msg *Message) {
	if client.RoomID == "" {
		client.SendError("NOT_IN_ROOM", "Not in any room")
		return
	}
	
	err := h.roomManager.LeaveRoom(client.RoomID, client.Player.ID)
	if err != nil {
		client.SendError("LEAVE_FAILED", err.Error())
		return
	}
	
	roomID := client.RoomID
	client.RoomID = ""
	
	// Broadcast room state
	h.broadcastRoomState(roomID)
	
	// Send updated room list to client
	h.handleListRooms(client, msg)
}

// handleAddBot handles adding a bot to a room
func (h *Hub) handleAddBot(client *Client, msg *Message) {
	if client.RoomID == "" {
		client.SendError("NOT_IN_ROOM", "Not in any room")
		return
	}
	
	var data AddBotData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid add bot data")
		return
	}
	
	err := h.roomManager.AddBot(client.RoomID, data.BotName, data.Difficulty, client.Player.ID)
	if err != nil {
		client.SendError("ADD_BOT_FAILED", err.Error())
		return
	}
	
	// Broadcast room state
	h.broadcastRoomState(client.RoomID)
}

// handlePlaceTile handles tile placement
func (h *Hub) handlePlaceTile(client *Client, msg *Message) {
	if client.RoomID == "" {
		client.SendError("NOT_IN_ROOM", "Not in any room")
		return
	}
	
	var data PlaceTileData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid place tile data")
		return
	}
	
	room, err := h.roomManager.GetRoom(client.RoomID)
	if err != nil {
		client.SendError("ROOM_NOT_FOUND", "Room not found")
		return
	}
	
	err = room.PlaceTile(client.Player.ID, data.Position, data.Rotation)
	if err != nil {
		client.SendError("PLACE_TILE_FAILED", err.Error())
		return
	}
	
	// Broadcast game state
	h.broadcastGameState(client.RoomID)
}

// handlePlaceMeeple handles meeple placement
func (h *Hub) handlePlaceMeeple(client *Client, msg *Message) {
	if client.RoomID == "" {
		client.SendError("NOT_IN_ROOM", "Not in any room")
		return
	}
	
	var data PlaceMeepleData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid place meeple data")
		return
	}
	
	room, err := h.roomManager.GetRoom(client.RoomID)
	if err != nil {
		client.SendError("ROOM_NOT_FOUND", "Room not found")
		return
	}
	
	err = room.PlaceMeeple(client.Player.ID, data.FeatureID)
	if err != nil {
		client.SendError("PLACE_MEEPLE_FAILED", err.Error())
		return
	}
	
	// End turn and broadcast state
	room.NextTurn()
	h.broadcastGameState(client.RoomID)
	h.sendTurnStart(client.RoomID)
}

// handlePing handles ping messages for latency calculation
func (h *Hub) handlePing(client *Client, msg *Message) {
	var data PingData
	if err := ParseMessage(msg, &data); err != nil {
		client.SendError("INVALID_DATA", "Invalid ping data")
		return
	}
	
	// Create pong response with original timestamp
	pongMsg, err := NewPongMessage(data.Timestamp, data.ClientID)
	if err != nil {
		log.Printf("Error creating pong message: %v", err)
		return
	}
	
	// Send pong response back to client
	client.SendMessage(pongMsg)
	
	log.Printf("Ping/Pong: Client %s latency measurement", data.ClientID)
}

// sendRoomState sends room state to a specific client
func (h *Hub) sendRoomState(client *Client, room *room.Room) {
	players := room.GetPlayers()
	
	msg, err := NewRoomStateMessage(room.ID, players, room.GameStarted, room.GameEnded)
	if err != nil {
		log.Printf("Error creating room state message: %v", err)
		return
	}
	
	client.SendMessage(msg)
}

// broadcastRoomState broadcasts room state to all clients in a room
func (h *Hub) broadcastRoomState(roomID string) {
	room, err := h.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}
	
	players := room.GetPlayers()
	msg, err := NewRoomStateMessage(roomID, players, room.GameStarted, room.GameEnded)
	if err != nil {
		log.Printf("Error creating room state message: %v", err)
		return
	}
	
	h.broadcastToRoom(roomID, msg)
}

// broadcastGameState broadcasts game state to all clients in a room
func (h *Hub) broadcastGameState(roomID string) {
	room, err := h.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}
	
	gameState := room.GetGameState()
	msg, err := NewGameStateMessage(gameState)
	if err != nil {
		log.Printf("Error creating game state message: %v", err)
		return
	}
	
	h.broadcastToRoom(roomID, msg)
}

// sendTurnStart sends turn start message to all clients in a room
func (h *Hub) sendTurnStart(roomID string) {
	room, err := h.roomManager.GetRoom(roomID)
	if err != nil {
		return
	}
	
	currentPlayer := room.GetCurrentPlayer()
	if currentPlayer == nil {
		return
	}
	
	gameState := room.GetGameState()
	validPlacements := room.GetValidPlacements()
	
	msg, err := NewTurnStartMessage(currentPlayer.ID, gameState.CurrentTile, validPlacements)
	if err != nil {
		log.Printf("Error creating turn start message: %v", err)
		return
	}
	
	h.broadcastToRoom(roomID, msg)
}

// broadcastToRoom broadcasts a message to all clients in a specific room
func (h *Hub) broadcastToRoom(roomID string, msg *Message) {
	for client := range h.clients {
		if client.RoomID == roomID {
			client.SendMessage(msg)
		}
	}
}

// processBotMoves processes bot moves periodically
func (h *Hub) processBotMoves() {
	for range h.botTicker.C {
		rooms := h.roomManager.ListRooms()
		
		for _, roomInfo := range rooms {
			if !roomInfo.GameStarted {
				continue
			}
			
			room, err := h.roomManager.GetRoom(roomInfo.ID)
			if err != nil {
				continue
			}
			
			if room.IsCurrentPlayerBot() {
				move, err := room.ProcessBotTurn()
				if err != nil {
					log.Printf("Error processing bot turn: %v", err)
					continue
				}
				
				// Broadcast the bot's move
				room.NextTurn()
				h.broadcastGameState(room.ID)
				h.sendTurnStart(room.ID)
				
				log.Printf("Bot made move: %+v", move)
			}
		}
	}
}

// StartGame starts a game in a room
func (h *Hub) StartGame(roomID, playerID string) error {
	err := h.roomManager.StartGame(roomID, playerID)
	if err != nil {
		return err
	}
	
	// Broadcast game start
	room, _ := h.roomManager.GetRoom(roomID)
	players := room.GetPlayers()
	
	msg, err := CreateMessage(MessageGameStart, GameStartData{
		RoomID:  roomID,
		Players: players,
	})
	if err != nil {
		return err
	}
	
	h.broadcastToRoom(roomID, msg)
	h.broadcastGameState(roomID)
	h.sendTurnStart(roomID)
	
	return nil
}
