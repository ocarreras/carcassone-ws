package game

import "fmt"

// TileEdge represents the type of edge on a tile
type TileEdge int

const (
	Road TileEdge = iota
	City
	Field
)

// Tile represents a Carcassonne tile with its edges and features
type Tile struct {
	ID       int
	North    TileEdge
	East     TileEdge
	South    TileEdge
	West     TileEdge
	Features []Feature
	HasShield bool // For cities
	HasMonastery bool
}

// Feature represents a feature on a tile (road, city, monastery, field)
type Feature struct {
	Type     FeatureType
	Edges    []Direction // Which edges this feature touches
	ID       int         // Unique identifier for scoring
	HasShield bool       // For cities
}

type FeatureType int

const (
	RoadFeature FeatureType = iota
	CityFeature
	MonasteryFeature
	FieldFeature
)

type Direction int

const (
	North Direction = iota
	East
	South
	West
)

// Position represents a position on the board
type Position struct {
	X, Y int
}

// PlacedTile represents a tile that has been placed on the board
type PlacedTile struct {
	Tile     *Tile
	Position Position
	Rotation int // 0, 90, 180, 270 degrees
	Meeples  []PlacedMeeple
}

// PlacedMeeple represents a meeple placed on a tile
type PlacedMeeple struct {
	PlayerID   string
	FeatureID  int
	Color      string
}

// Rotate rotates the tile by 90 degrees clockwise
func (t *Tile) Rotate() {
	t.North, t.East, t.South, t.West = t.West, t.North, t.East, t.South
}

// GetEdge returns the edge in the specified direction after rotation
func (pt *PlacedTile) GetEdge(dir Direction) TileEdge {
	rotatedDir := (int(dir) - pt.Rotation/90 + 4) % 4
	switch Direction(rotatedDir) {
	case North:
		return pt.Tile.North
	case East:
		return pt.Tile.East
	case South:
		return pt.Tile.South
	case West:
		return pt.Tile.West
	default:
		return Field
	}
}

// CanPlaceAt checks if a tile can be placed at the given position
func (pt *PlacedTile) CanPlaceAt(board map[Position]*PlacedTile, pos Position) bool {
	// Check if position is already occupied
	if _, exists := board[pos]; exists {
		return false
	}

	// Check if there's at least one adjacent tile
	hasAdjacent := false
	adjacentPositions := []struct {
		pos Position
		dir Direction
		opposite Direction
	}{
		{Position{pos.X, pos.Y - 1}, North, South},
		{Position{pos.X + 1, pos.Y}, East, West},
		{Position{pos.X, pos.Y + 1}, South, North},
		{Position{pos.X - 1, pos.Y}, West, East},
	}

	for _, adj := range adjacentPositions {
		if adjacentTile, exists := board[adj.pos]; exists {
			hasAdjacent = true
			// Check if edges match
			myEdge := pt.GetEdge(adj.dir)
			theirEdge := adjacentTile.GetEdge(adj.opposite)
			if myEdge != theirEdge {
				return false
			}
		}
	}

	return hasAdjacent
}

// CreateStandardTileSet creates the standard 72-tile Carcassonne set
func CreateStandardTileSet() []*Tile {
	tiles := make([]*Tile, 0, 72)
	
	// Starting tile (monastery with road)
	tiles = append(tiles, &Tile{
		ID: 0,
		North: Field, East: Road, South: Road, West: Field,
		HasMonastery: true,
		Features: []Feature{
			{Type: MonasteryFeature, Edges: []Direction{}, ID: 0},
			{Type: RoadFeature, Edges: []Direction{East, South}, ID: 1},
		},
	})

	// Add more tiles - this is a simplified set for demonstration
	// In a full implementation, you'd add all 72 unique tiles
	
	// Road tiles
	for i := 1; i <= 8; i++ {
		tiles = append(tiles, &Tile{
			ID: i,
			North: Field, East: Road, South: Field, West: Road,
			Features: []Feature{
				{Type: RoadFeature, Edges: []Direction{East, West}, ID: 0},
				{Type: FieldFeature, Edges: []Direction{North, South}, ID: 1},
			},
		})
	}

	// City tiles
	for i := 9; i <= 16; i++ {
		tiles = append(tiles, &Tile{
			ID: i,
			North: City, East: Field, South: Field, West: Field,
			Features: []Feature{
				{Type: CityFeature, Edges: []Direction{North}, ID: 0},
				{Type: FieldFeature, Edges: []Direction{East, South, West}, ID: 1},
			},
		})
	}

	// Monastery tiles
	for i := 17; i <= 20; i++ {
		tiles = append(tiles, &Tile{
			ID: i,
			North: Field, East: Field, South: Field, West: Field,
			HasMonastery: true,
			Features: []Feature{
				{Type: MonasteryFeature, Edges: []Direction{}, ID: 0},
				{Type: FieldFeature, Edges: []Direction{North, East, South, West}, ID: 1},
			},
		})
	}

	return tiles
}

func (t *Tile) String() string {
	return fmt.Sprintf("Tile{ID:%d, N:%v, E:%v, S:%v, W:%v, Monastery:%v}", 
		t.ID, t.North, t.East, t.South, t.West, t.HasMonastery)
}
