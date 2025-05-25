package game

import (
	"fmt"
	"math/rand"
	"time"
)

// Board represents the game board
type Board struct {
	Tiles        map[Position]*PlacedTile
	TileDeck     []*Tile
	CurrentTile  *Tile
	Players      []*Player
	CurrentPlayer int
	GameStarted  bool
	GameEnded    bool
	Scores       map[string]int
}

// Player represents a player in the game
type Player struct {
	ID       string
	Name     string
	Color    string
	Meeples  int
	IsBot    bool
	Score    int
}

// NewBoard creates a new game board
func NewBoard() *Board {
	tiles := CreateStandardTileSet()
	
	// Shuffle the deck
	rand.Seed(time.Now().UnixNano())
	for i := len(tiles) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		tiles[i], tiles[j] = tiles[j], tiles[i]
	}

	board := &Board{
		Tiles:    make(map[Position]*PlacedTile),
		TileDeck: tiles[1:], // Skip the starting tile
		Players:  make([]*Player, 0),
		Scores:   make(map[string]int),
	}

	// Place the starting tile at (0, 0)
	startingTile := &PlacedTile{
		Tile:     tiles[0],
		Position: Position{X: 0, Y: 0},
		Rotation: 0,
		Meeples:  make([]PlacedMeeple, 0),
	}
	board.Tiles[Position{X: 0, Y: 0}] = startingTile

	return board
}

// AddPlayer adds a player to the game
func (b *Board) AddPlayer(player *Player) error {
	if len(b.Players) >= 5 {
		return fmt.Errorf("maximum 5 players allowed")
	}
	
	if b.GameStarted {
		return fmt.Errorf("game already started")
	}

	player.Meeples = 7 // Each player starts with 7 meeples
	player.Score = 0
	b.Players = append(b.Players, player)
	b.Scores[player.ID] = 0
	
	return nil
}

// StartGame starts the game
func (b *Board) StartGame() error {
	if len(b.Players) < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}
	
	if b.GameStarted {
		return fmt.Errorf("game already started")
	}

	b.GameStarted = true
	b.CurrentPlayer = 0
	b.DrawNextTile()
	
	return nil
}

// DrawNextTile draws the next tile from the deck
func (b *Board) DrawNextTile() bool {
	if len(b.TileDeck) == 0 {
		b.GameEnded = true
		return false
	}
	
	b.CurrentTile = b.TileDeck[0]
	b.TileDeck = b.TileDeck[1:]
	return true
}

// GetValidPlacements returns all valid positions and rotations for the current tile
func (b *Board) GetValidPlacements() []PlacementOption {
	if b.CurrentTile == nil {
		return nil
	}

	validPlacements := make([]PlacementOption, 0)
	
	// Get all possible positions (adjacent to existing tiles)
	possiblePositions := b.getPossiblePositions()
	
	for _, pos := range possiblePositions {
		for rotation := 0; rotation < 360; rotation += 90 {
			placedTile := &PlacedTile{
				Tile:     b.CurrentTile,
				Position: pos,
				Rotation: rotation,
			}
			
			if placedTile.CanPlaceAt(b.Tiles, pos) {
				validPlacements = append(validPlacements, PlacementOption{
					Position: pos,
					Rotation: rotation,
				})
			}
		}
	}
	
	return validPlacements
}

// PlacementOption represents a valid tile placement
type PlacementOption struct {
	Position Position
	Rotation int
}

// getPossiblePositions returns all positions adjacent to existing tiles
func (b *Board) getPossiblePositions() []Position {
	positions := make(map[Position]bool)
	
	for pos := range b.Tiles {
		// Add all adjacent positions
		adjacent := []Position{
			{pos.X, pos.Y - 1}, // North
			{pos.X + 1, pos.Y}, // East
			{pos.X, pos.Y + 1}, // South
			{pos.X - 1, pos.Y}, // West
		}
		
		for _, adjPos := range adjacent {
			if _, exists := b.Tiles[adjPos]; !exists {
				positions[adjPos] = true
			}
		}
	}
	
	result := make([]Position, 0, len(positions))
	for pos := range positions {
		result = append(result, pos)
	}
	
	return result
}

// PlaceTile places a tile on the board
func (b *Board) PlaceTile(pos Position, rotation int) error {
	if b.CurrentTile == nil {
		return fmt.Errorf("no current tile to place")
	}
	
	placedTile := &PlacedTile{
		Tile:     b.CurrentTile,
		Position: pos,
		Rotation: rotation,
		Meeples:  make([]PlacedMeeple, 0),
	}
	
	if !placedTile.CanPlaceAt(b.Tiles, pos) {
		return fmt.Errorf("invalid tile placement")
	}
	
	b.Tiles[pos] = placedTile
	b.CurrentTile = nil
	
	return nil
}

// PlaceMeeple places a meeple on the last placed tile
func (b *Board) PlaceMeeple(playerID string, featureID int) error {
	player := b.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	
	if player.Meeples <= 0 {
		return fmt.Errorf("no meeples available")
	}
	
	// Find the last placed tile (this is simplified - in a real game you'd track this better)
	var lastTile *PlacedTile
	for _, tile := range b.Tiles {
		if len(tile.Meeples) == 0 { // Assuming the last placed tile has no meeples yet
			lastTile = tile
			break
		}
	}
	
	if lastTile == nil {
		return fmt.Errorf("no tile to place meeple on")
	}
	
	// Check if feature is valid and not already occupied
	if featureID >= len(lastTile.Tile.Features) {
		return fmt.Errorf("invalid feature ID")
	}
	
	// Check if feature is already occupied (simplified check)
	for _, meeple := range lastTile.Meeples {
		if meeple.FeatureID == featureID {
			return fmt.Errorf("feature already occupied")
		}
	}
	
	meeple := PlacedMeeple{
		PlayerID:  playerID,
		FeatureID: featureID,
		Color:     player.Color,
	}
	
	lastTile.Meeples = append(lastTile.Meeples, meeple)
	player.Meeples--
	
	return nil
}

// NextTurn advances to the next player's turn
func (b *Board) NextTurn() {
	b.CurrentPlayer = (b.CurrentPlayer + 1) % len(b.Players)
	if !b.DrawNextTile() {
		b.EndGame()
	}
}

// GetCurrentPlayer returns the current player
func (b *Board) GetCurrentPlayer() *Player {
	if len(b.Players) == 0 {
		return nil
	}
	return b.Players[b.CurrentPlayer]
}

// GetPlayer returns a player by ID
func (b *Board) GetPlayer(playerID string) *Player {
	for _, player := range b.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}

// EndGame ends the game and calculates final scores
func (b *Board) EndGame() {
	b.GameEnded = true
	// Calculate final scores for incomplete features
	b.calculateFinalScores()
}

// calculateFinalScores calculates final scores for incomplete features
func (b *Board) calculateFinalScores() {
	// This is a simplified implementation
	// In a full game, you'd need to:
	// 1. Find all incomplete roads, cities, and monasteries
	// 2. Score them at reduced points
	// 3. Score farms based on completed cities they supply
	
	for _, player := range b.Players {
		b.Scores[player.ID] = player.Score
	}
}

// GetGameState returns the current game state
func (b *Board) GetGameState() GameState {
	return GameState{
		Tiles:         b.Tiles,
		CurrentTile:   b.CurrentTile,
		Players:       b.Players,
		CurrentPlayer: b.CurrentPlayer,
		GameStarted:   b.GameStarted,
		GameEnded:     b.GameEnded,
		Scores:        b.Scores,
		TilesLeft:     len(b.TileDeck),
	}
}

// GameState represents the current state of the game
type GameState struct {
	Tiles         map[Position]*PlacedTile `json:"tiles"`
	CurrentTile   *Tile                    `json:"currentTile"`
	Players       []*Player                `json:"players"`
	CurrentPlayer int                      `json:"currentPlayer"`
	GameStarted   bool                     `json:"gameStarted"`
	GameEnded     bool                     `json:"gameEnded"`
	Scores        map[string]int           `json:"scores"`
	TilesLeft     int                      `json:"tilesLeft"`
}
