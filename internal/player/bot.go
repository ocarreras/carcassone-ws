package player

import (
	"math/rand"
	"time"
	"carcassonne-ws/internal/game"
)

// Bot represents an AI player
type Bot struct {
	Player *game.Player
	Difficulty string // "easy", "medium", "hard"
}

// NewBot creates a new bot player
func NewBot(id, name, color string) *Bot {
	return &Bot{
		Player: &game.Player{
			ID:      id,
			Name:    name,
			Color:   color,
			Meeples: 7,
			IsBot:   true,
			Score:   0,
		},
		Difficulty: "easy",
	}
}

// MakeMove makes a move for the bot
func (b *Bot) MakeMove(board *game.Board) (BotMove, error) {
	rand.Seed(time.Now().UnixNano())
	
	// Get valid placements
	validPlacements := board.GetValidPlacements()
	if len(validPlacements) == 0 {
		return BotMove{}, nil // No valid moves
	}
	
	// Choose a random valid placement
	placement := validPlacements[rand.Intn(len(validPlacements))]
	
	move := BotMove{
		TilePlacement: TilePlacement{
			Position: placement.Position,
			Rotation: placement.Rotation,
		},
	}
	
	// Decide whether to place a meeple (50% chance for easy bots)
	if b.Player.Meeples > 0 && rand.Float32() < 0.5 {
		// Choose a random feature to place meeple on
		if board.CurrentTile != nil && len(board.CurrentTile.Features) > 0 {
			featureID := rand.Intn(len(board.CurrentTile.Features))
			move.MeeplePlacement = &MeeplePlacement{
				FeatureID: featureID,
			}
		}
	}
	
	return move, nil
}

// BotMove represents a complete move by a bot
type BotMove struct {
	TilePlacement   TilePlacement    `json:"tilePlacement"`
	MeeplePlacement *MeeplePlacement `json:"meeplePlacement,omitempty"`
}

// TilePlacement represents a tile placement
type TilePlacement struct {
	Position game.Position `json:"position"`
	Rotation int           `json:"rotation"`
}

// MeeplePlacement represents a meeple placement
type MeeplePlacement struct {
	FeatureID int `json:"featureId"`
}

// ExecuteMove executes the bot's move on the board
func (b *Bot) ExecuteMove(board *game.Board, move BotMove) error {
	// Place the tile
	err := board.PlaceTile(move.TilePlacement.Position, move.TilePlacement.Rotation)
	if err != nil {
		return err
	}
	
	// Place meeple if specified
	if move.MeeplePlacement != nil {
		err = board.PlaceMeeple(b.Player.ID, move.MeeplePlacement.FeatureID)
		if err != nil {
			// Meeple placement failed, but tile placement succeeded
			// This is okay, just continue without the meeple
		}
	}
	
	return nil
}

// GetStrategy returns the bot's current strategy
func (b *Bot) GetStrategy() string {
	switch b.Difficulty {
	case "easy":
		return "Random valid moves"
	case "medium":
		return "Prioritize completing features"
	case "hard":
		return "Advanced scoring optimization"
	default:
		return "Random valid moves"
	}
}

// SetDifficulty sets the bot's difficulty level
func (b *Bot) SetDifficulty(difficulty string) {
	b.Difficulty = difficulty
}

// ShouldPlaceMeeple determines if the bot should place a meeple
func (b *Bot) ShouldPlaceMeeple(board *game.Board, tile *game.Tile) (bool, int) {
	if b.Player.Meeples <= 0 {
		return false, -1
	}
	
	switch b.Difficulty {
	case "easy":
		// 50% chance to place meeple on random feature
		if rand.Float32() < 0.5 && len(tile.Features) > 0 {
			return true, rand.Intn(len(tile.Features))
		}
		return false, -1
		
	case "medium":
		// Prefer monasteries and short roads/cities
		for i, feature := range tile.Features {
			if feature.Type == game.MonasteryFeature {
				return true, i
			}
		}
		// Fallback to random
		if rand.Float32() < 0.3 && len(tile.Features) > 0 {
			return true, rand.Intn(len(tile.Features))
		}
		return false, -1
		
	case "hard":
		// More sophisticated strategy would go here
		// For now, same as medium
		return b.ShouldPlaceMeeple(board, tile)
		
	default:
		return false, -1
	}
}

// ChooseBestPlacement chooses the best tile placement based on difficulty
func (b *Bot) ChooseBestPlacement(validPlacements []game.PlacementOption, board *game.Board) game.PlacementOption {
	if len(validPlacements) == 0 {
		return game.PlacementOption{}
	}
	
	switch b.Difficulty {
	case "easy":
		// Random placement
		return validPlacements[rand.Intn(len(validPlacements))]
		
	case "medium":
		// Prefer placements that complete features or extend existing ones
		// For now, just random (would implement scoring logic here)
		return validPlacements[rand.Intn(len(validPlacements))]
		
	case "hard":
		// Advanced placement strategy
		// For now, just random (would implement advanced AI here)
		return validPlacements[rand.Intn(len(validPlacements))]
		
	default:
		return validPlacements[rand.Intn(len(validPlacements))]
	}
}
