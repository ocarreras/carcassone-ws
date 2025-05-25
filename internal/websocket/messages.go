package websocket

import (
	"encoding/json"
	"time"
	"carcassonne-ws/internal/game"
	"carcassonne-ws/internal/player"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Connection & Room Management
	MessageConnect    MessageType = "CONNECT"
	MessageListRooms  MessageType = "LIST_ROOMS"
	MessageCreateRoom MessageType = "CREATE_ROOM"
	MessageJoinRoom   MessageType = "JOIN_ROOM"
	MessageLeaveRoom  MessageType = "LEAVE_ROOM"
	MessageAddBot     MessageType = "ADD_BOT"
	
	// Game Flow
	MessageGameStart MessageType = "GAME_START"
	MessageTurnStart MessageType = "TURN_START"
	MessagePlaceTile MessageType = "PLACE_TILE"
	MessagePlaceMeeple MessageType = "PLACE_MEEPLE"
	MessageTurnEnd   MessageType = "TURN_END"
	MessageGameEnd   MessageType = "GAME_END"
	
	// State Synchronization
	MessageRoomState   MessageType = "ROOM_STATE"
	MessageGameState   MessageType = "GAME_STATE"
	MessagePlayerUpdate MessageType = "PLAYER_UPDATE"
	
	// Error handling
	MessageError MessageType = "ERROR"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	MessageID string          `json:"messageId"`
}

// ConnectData represents connection message data
type ConnectData struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Color    string `json:"color"`
}

// ListRoomsData represents list rooms response data
type ListRoomsData struct {
	Rooms []RoomInfo `json:"rooms"`
}

// RoomInfo represents room information
type RoomInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	GameStarted bool   `json:"gameStarted"`
	CreatedBy   string `json:"createdBy"`
}

// CreateRoomData represents create room message data
type CreateRoomData struct {
	RoomName   string `json:"roomName"`
	MaxPlayers int    `json:"maxPlayers"`
}

// JoinRoomData represents join room message data
type JoinRoomData struct {
	RoomID string `json:"roomId"`
}

// LeaveRoomData represents leave room message data
type LeaveRoomData struct {
	RoomID string `json:"roomId"`
}

// AddBotData represents add bot message data
type AddBotData struct {
	BotName    string `json:"botName"`
	Difficulty string `json:"difficulty"`
}

// GameStartData represents game start message data
type GameStartData struct {
	RoomID  string         `json:"roomId"`
	Players []*game.Player `json:"players"`
}

// TurnStartData represents turn start message data
type TurnStartData struct {
	CurrentPlayer string     `json:"currentPlayer"`
	CurrentTile   *game.Tile `json:"currentTile"`
	ValidPlacements []game.PlacementOption `json:"validPlacements"`
}

// PlaceTileData represents place tile message data
type PlaceTileData struct {
	Position game.Position `json:"position"`
	Rotation int           `json:"rotation"`
}

// PlaceMeepleData represents place meeple message data
type PlaceMeepleData struct {
	FeatureID int `json:"featureId"`
}

// TurnEndData represents turn end message data
type TurnEndData struct {
	PlayerID    string            `json:"playerId"`
	ScoreChange int               `json:"scoreChange"`
	NextPlayer  string            `json:"nextPlayer"`
	GameState   game.GameState    `json:"gameState"`
}

// GameEndData represents game end message data
type GameEndData struct {
	Winner     string         `json:"winner"`
	FinalScore map[string]int `json:"finalScore"`
	GameState  game.GameState `json:"gameState"`
}

// RoomStateData represents room state message data
type RoomStateData struct {
	RoomID      string         `json:"roomId"`
	Players     []*game.Player `json:"players"`
	GameStarted bool           `json:"gameStarted"`
	GameEnded   bool           `json:"gameEnded"`
}

// GameStateData represents game state message data
type GameStateData struct {
	GameState game.GameState `json:"gameState"`
}

// PlayerUpdateData represents player update message data
type PlayerUpdateData struct {
	Player *game.Player `json:"player"`
}

// ErrorData represents error message data
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BotMoveData represents bot move message data
type BotMoveData struct {
	BotID string           `json:"botId"`
	Move  player.BotMove   `json:"move"`
}

// CreateMessage creates a new message with the given type and data
func CreateMessage(msgType MessageType, data interface{}) (*Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	
	return &Message{
		Type:      msgType,
		Data:      dataBytes,
		Timestamp: time.Now(),
		MessageID: generateMessageID(),
	}, nil
}

// ParseMessage parses a message and returns the typed data
func ParseMessage(msg *Message, target interface{}) error {
	return json.Unmarshal(msg.Data, target)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // Simplified for demo
	}
	return string(b)
}

// MessageBuilder provides a fluent interface for building messages
type MessageBuilder struct {
	msgType MessageType
	data    interface{}
}

// NewMessageBuilder creates a new message builder
func NewMessageBuilder(msgType MessageType) *MessageBuilder {
	return &MessageBuilder{msgType: msgType}
}

// WithData sets the message data
func (mb *MessageBuilder) WithData(data interface{}) *MessageBuilder {
	mb.data = data
	return mb
}

// Build builds the message
func (mb *MessageBuilder) Build() (*Message, error) {
	return CreateMessage(mb.msgType, mb.data)
}

// Common message builders for convenience
func NewConnectMessage(playerID, name, color string) (*Message, error) {
	return CreateMessage(MessageConnect, ConnectData{
		PlayerID: playerID,
		Name:     name,
		Color:    color,
	})
}

func NewErrorMessage(code, message string) (*Message, error) {
	return CreateMessage(MessageError, ErrorData{
		Code:    code,
		Message: message,
	})
}

func NewGameStateMessage(gameState game.GameState) (*Message, error) {
	return CreateMessage(MessageGameState, GameStateData{
		GameState: gameState,
	})
}

func NewRoomStateMessage(roomID string, players []*game.Player, gameStarted, gameEnded bool) (*Message, error) {
	return CreateMessage(MessageRoomState, RoomStateData{
		RoomID:      roomID,
		Players:     players,
		GameStarted: gameStarted,
		GameEnded:   gameEnded,
	})
}

func NewTurnStartMessage(currentPlayer string, currentTile *game.Tile, validPlacements []game.PlacementOption) (*Message, error) {
	return CreateMessage(MessageTurnStart, TurnStartData{
		CurrentPlayer:   currentPlayer,
		CurrentTile:     currentTile,
		ValidPlacements: validPlacements,
	})
}
