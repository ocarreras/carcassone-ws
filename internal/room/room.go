package room

import (
	"fmt"
	"sync"
	"time"
	"carcassonne-ws/internal/game"
	"carcassonne-ws/internal/player"
	"github.com/google/uuid"
)

// Room represents a game room
type Room struct {
	ID          string
	Name        string
	MaxPlayers  int
	CreatedBy   string
	CreatedAt   time.Time
	Players     map[string]*game.Player
	Bots        map[string]*player.Bot
	Board       *game.Board
	GameStarted bool
	GameEnded   bool
	mutex       sync.RWMutex
}

// NewRoom creates a new game room
func NewRoom(name, createdBy string, maxPlayers int) *Room {
	if maxPlayers < 2 || maxPlayers > 5 {
		maxPlayers = 5
	}
	
	return &Room{
		ID:         uuid.New().String(),
		Name:       name,
		MaxPlayers: maxPlayers,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
		Players:    make(map[string]*game.Player),
		Bots:       make(map[string]*player.Bot),
		Board:      game.NewBoard(),
	}
}

// AddPlayer adds a player to the room
func (r *Room) AddPlayer(player *game.Player) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.GameStarted {
		return fmt.Errorf("game already started")
	}
	
	if len(r.Players)+len(r.Bots) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}
	
	if _, exists := r.Players[player.ID]; exists {
		return fmt.Errorf("player already in room")
	}
	
	r.Players[player.ID] = player
	r.Board.AddPlayer(player)
	
	return nil
}

// RemovePlayer removes a player from the room
func (r *Room) RemovePlayer(playerID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.GameStarted {
		return fmt.Errorf("cannot leave during game")
	}
	
	if _, exists := r.Players[playerID]; !exists {
		return fmt.Errorf("player not in room")
	}
	
	delete(r.Players, playerID)
	
	// Remove from board players
	for i, p := range r.Board.Players {
		if p.ID == playerID {
			r.Board.Players = append(r.Board.Players[:i], r.Board.Players[i+1:]...)
			break
		}
	}
	
	return nil
}

// AddBot adds a bot to the room
func (r *Room) AddBot(botName, difficulty, creatorID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.GameStarted {
		return fmt.Errorf("game already started")
	}
	
	if r.CreatedBy != creatorID {
		return fmt.Errorf("only room creator can add bots")
	}
	
	if len(r.Players)+len(r.Bots) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}
	
	botID := uuid.New().String()
	colors := []string{"red", "blue", "green", "yellow", "black"}
	usedColors := make(map[string]bool)
	
	// Track used colors
	for _, p := range r.Players {
		usedColors[p.Color] = true
	}
	for _, b := range r.Bots {
		usedColors[b.Player.Color] = true
	}
	
	// Find available color
	var botColor string
	for _, color := range colors {
		if !usedColors[color] {
			botColor = color
			break
		}
	}
	
	if botColor == "" {
		return fmt.Errorf("no available colors for bot")
	}
	
	bot := player.NewBot(botID, botName, botColor)
	bot.SetDifficulty(difficulty)
	
	r.Bots[botID] = bot
	r.Board.AddPlayer(bot.Player)
	
	return nil
}

// StartGame starts the game in the room
func (r *Room) StartGame() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.GameStarted {
		return fmt.Errorf("game already started")
	}
	
	totalPlayers := len(r.Players) + len(r.Bots)
	if totalPlayers < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}
	
	err := r.Board.StartGame()
	if err != nil {
		return err
	}
	
	r.GameStarted = true
	return nil
}

// GetPlayers returns all players (human and bot) in the room
func (r *Room) GetPlayers() []*game.Player {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	players := make([]*game.Player, 0, len(r.Players)+len(r.Bots))
	
	for _, player := range r.Players {
		players = append(players, player)
	}
	
	for _, bot := range r.Bots {
		players = append(players, bot.Player)
	}
	
	return players
}

// GetPlayerCount returns the total number of players in the room
func (r *Room) GetPlayerCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return len(r.Players) + len(r.Bots)
}

// IsCreator checks if the given player ID is the room creator
func (r *Room) IsCreator(playerID string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.CreatedBy == playerID
}

// CanStart checks if the game can be started
func (r *Room) CanStart() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return !r.GameStarted && len(r.Players)+len(r.Bots) >= 2
}

// GetBot returns a bot by ID
func (r *Room) GetBot(botID string) *player.Bot {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.Bots[botID]
}

// GetCurrentPlayer returns the current player
func (r *Room) GetCurrentPlayer() *game.Player {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.Board.GetCurrentPlayer()
}

// IsCurrentPlayerBot checks if the current player is a bot
func (r *Room) IsCurrentPlayerBot() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	currentPlayer := r.Board.GetCurrentPlayer()
	if currentPlayer == nil {
		return false
	}
	
	_, isBot := r.Bots[currentPlayer.ID]
	return isBot
}

// ProcessBotTurn processes a bot's turn
func (r *Room) ProcessBotTurn() (*player.BotMove, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	currentPlayer := r.Board.GetCurrentPlayer()
	if currentPlayer == nil {
		return nil, fmt.Errorf("no current player")
	}
	
	bot, exists := r.Bots[currentPlayer.ID]
	if !exists {
		return nil, fmt.Errorf("current player is not a bot")
	}
	
	move, err := bot.MakeMove(r.Board)
	if err != nil {
		return nil, err
	}
	
	err = bot.ExecuteMove(r.Board, move)
	if err != nil {
		return nil, err
	}
	
	return &move, nil
}

// PlaceTile places a tile on the board
func (r *Room) PlaceTile(playerID string, pos game.Position, rotation int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	currentPlayer := r.Board.GetCurrentPlayer()
	if currentPlayer == nil || currentPlayer.ID != playerID {
		return fmt.Errorf("not your turn")
	}
	
	return r.Board.PlaceTile(pos, rotation)
}

// PlaceMeeple places a meeple on the board
func (r *Room) PlaceMeeple(playerID string, featureID int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	currentPlayer := r.Board.GetCurrentPlayer()
	if currentPlayer == nil || currentPlayer.ID != playerID {
		return fmt.Errorf("not your turn")
	}
	
	return r.Board.PlaceMeeple(playerID, featureID)
}

// NextTurn advances to the next turn
func (r *Room) NextTurn() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.Board.NextTurn()
	
	if r.Board.GameEnded {
		r.GameEnded = true
	}
}

// GetGameState returns the current game state
func (r *Room) GetGameState() game.GameState {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.Board.GetGameState()
}

// GetValidPlacements returns valid placements for the current tile
func (r *Room) GetValidPlacements() []game.PlacementOption {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return r.Board.GetValidPlacements()
}

// GetRoomInfo returns room information
func (r *Room) GetRoomInfo() RoomInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	return RoomInfo{
		ID:          r.ID,
		Name:        r.Name,
		PlayerCount: len(r.Players) + len(r.Bots),
		MaxPlayers:  r.MaxPlayers,
		GameStarted: r.GameStarted,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
	}
}

// RoomInfo represents room information for listing
type RoomInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PlayerCount int       `json:"playerCount"`
	MaxPlayers  int       `json:"maxPlayers"`
	GameStarted bool      `json:"gameStarted"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
}
