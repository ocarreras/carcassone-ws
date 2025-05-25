package room

import (
	"fmt"
	"sync"
	"carcassonne-ws/internal/game"
)

// Manager manages all game rooms
type Manager struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

// NewManager creates a new room manager
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new room
func (m *Manager) CreateRoom(name, createdBy string, maxPlayers int) (*Room, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	room := NewRoom(name, createdBy, maxPlayers)
	m.rooms[room.ID] = room
	
	return room, nil
}

// GetRoom returns a room by ID
func (m *Manager) GetRoom(roomID string) (*Room, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	room, exists := m.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}
	
	return room, nil
}

// JoinRoom adds a player to a room
func (m *Manager) JoinRoom(roomID string, player *game.Player) error {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}
	
	return room.AddPlayer(player)
}

// LeaveRoom removes a player from a room
func (m *Manager) LeaveRoom(roomID, playerID string) error {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}
	
	err = room.RemovePlayer(playerID)
	if err != nil {
		return err
	}
	
	// Clean up empty rooms
	if room.GetPlayerCount() == 0 && !room.GameStarted {
		m.mutex.Lock()
		delete(m.rooms, roomID)
		m.mutex.Unlock()
	}
	
	return nil
}

// ListRooms returns all available rooms
func (m *Manager) ListRooms() []RoomInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	rooms := make([]RoomInfo, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room.GetRoomInfo())
	}
	
	return rooms
}

// GetActiveRooms returns only rooms that are not full and not started
func (m *Manager) GetActiveRooms() []RoomInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	rooms := make([]RoomInfo, 0)
	for _, room := range m.rooms {
		info := room.GetRoomInfo()
		if !info.GameStarted && info.PlayerCount < info.MaxPlayers {
			rooms = append(rooms, info)
		}
	}
	
	return rooms
}

// FindPlayerRoom finds the room a player is currently in
func (m *Manager) FindPlayerRoom(playerID string) (*Room, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, room := range m.rooms {
		if _, exists := room.Players[playerID]; exists {
			return room, nil
		}
	}
	
	return nil, fmt.Errorf("player not in any room")
}

// CleanupEmptyRooms removes empty rooms that are not in progress
func (m *Manager) CleanupEmptyRooms() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for roomID, room := range m.rooms {
		if room.GetPlayerCount() == 0 && !room.GameStarted {
			delete(m.rooms, roomID)
		}
	}
}

// GetRoomCount returns the total number of rooms
func (m *Manager) GetRoomCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return len(m.rooms)
}

// GetTotalPlayers returns the total number of players across all rooms
func (m *Manager) GetTotalPlayers() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	total := 0
	for _, room := range m.rooms {
		total += room.GetPlayerCount()
	}
	
	return total
}

// StartGame starts a game in the specified room
func (m *Manager) StartGame(roomID, playerID string) error {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}
	
	if !room.IsCreator(playerID) {
		return fmt.Errorf("only room creator can start the game")
	}
	
	if !room.CanStart() {
		return fmt.Errorf("cannot start game: need at least 2 players")
	}
	
	return room.StartGame()
}

// AddBot adds a bot to the specified room
func (m *Manager) AddBot(roomID, botName, difficulty, creatorID string) error {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}
	
	return room.AddBot(botName, difficulty, creatorID)
}
