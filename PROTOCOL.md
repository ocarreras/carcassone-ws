# Carcassonne WebSocket Protocol Specification

Version: 1.0  
Date: 2025-01-01  
Protocol: WebSocket over TCP  

## Table of Contents

1. [Overview](#overview)
2. [Connection](#connection)
3. [Message Format](#message-format)
4. [Authentication & Session Management](#authentication--session-management)
5. [Room Management](#room-management)
6. [Game Flow](#game-flow)
7. [State Synchronization](#state-synchronization)
8. [Error Handling](#error-handling)
9. [Data Types](#data-types)
10. [Message Reference](#message-reference)
11. [Game Rules Implementation](#game-rules-implementation)
12. [Examples](#examples)

## Overview

The Carcassonne WebSocket Protocol enables real-time multiplayer gameplay for the Carcassonne board game. The protocol supports:

- Real-time bidirectional communication
- Room-based game sessions (2-5 players)
- AI bot integration
- Complete game state synchronization
- Turn-based gameplay with validation

### Protocol Characteristics

- **Transport**: WebSocket (RFC 6455)
- **Encoding**: JSON (UTF-8)
- **Message Pattern**: Request/Response + Server Push
- **State Management**: Server-authoritative
- **Concurrency**: Multiple concurrent games supported

## Connection

### Endpoint
```
ws://host:port/ws
```

### Connection Lifecycle

1. **WebSocket Handshake**: Standard WebSocket upgrade
2. **Authentication**: Send `CONNECT` message with player credentials
3. **Session Active**: Bidirectional message exchange
4. **Disconnection**: Graceful close or timeout

### Connection States

| State | Description |
|-------|-------------|
| `CONNECTING` | WebSocket handshake in progress |
| `AUTHENTICATED` | Player connected and authenticated |
| `IN_ROOM` | Player joined a game room |
| `IN_GAME` | Active game session |
| `DISCONNECTED` | Connection closed |

## Message Format

All messages follow a standardized JSON format:

```json
{
  "type": "MESSAGE_TYPE",
  "data": { /* message-specific payload */ },
  "timestamp": "2024-01-01T00:00:00Z",
  "messageId": "unique-identifier"
}
```

### Message Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | Message type identifier |
| `data` | object | Yes | Message payload (can be empty object) |
| `timestamp` | string | Yes | ISO 8601 timestamp |
| `messageId` | string | Yes | Unique message identifier |

### Message Types

Messages are categorized into functional groups:

- **Connection**: `CONNECT`
- **Room Management**: `LIST_ROOMS`, `CREATE_ROOM`, `JOIN_ROOM`, `LEAVE_ROOM`, `ADD_BOT`
- **Game Flow**: `GAME_START`, `TURN_START`, `PLACE_TILE`, `PLACE_MEEPLE`, `TURN_END`, `GAME_END`
- **State Sync**: `ROOM_STATE`, `GAME_STATE`, `PLAYER_UPDATE`
- **System**: `ERROR`, `PING`, `PONG`

## Authentication & Session Management

### Initial Connection

Client must send `CONNECT` message immediately after WebSocket connection:

```json
{
  "type": "CONNECT",
  "data": {
    "playerId": "unique-player-id",
    "name": "Player Name",
    "color": "red"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "messageId": "msg-001"
}
```

### Session Management

- **Player ID**: Unique identifier for reconnection
- **Session Timeout**: 5 minutes of inactivity
- **Reconnection**: Same `playerId` can reconnect to existing session
- **Cleanup**: Inactive sessions are automatically cleaned up

## Room Management

### Room Lifecycle

1. **Creation**: Player creates room and becomes host
2. **Joining**: Other players join the room
3. **Configuration**: Host adds bots, configures settings
4. **Game Start**: Host initiates game when ready
5. **Cleanup**: Room destroyed when empty or game ends

### Room States

| State | Description | Allowed Actions |
|-------|-------------|-----------------|
| `WAITING` | Room created, waiting for players | Join, Add Bot, Start Game |
| `STARTING` | Game initialization in progress | None |
| `IN_PROGRESS` | Active game session | Game actions only |
| `FINISHED` | Game completed | Leave room |

### Room Capacity

- **Minimum**: 2 players (human + bot combinations allowed)
- **Maximum**: 5 players
- **Host Privileges**: Only room creator can add bots and start games

## Game Flow

### Turn Structure

Each turn consists of:

1. **Tile Draw**: Server provides current tile to active player
2. **Tile Placement**: Player places tile on board
3. **Meeple Placement** (Optional): Player places meeple on placed tile
4. **Scoring**: Server calculates and updates scores
5. **Turn End**: Advance to next player

### Game Phases

| Phase | Description | Duration |
|-------|-------------|----------|
| `SETUP` | Initial game state preparation | Instant |
| `PLAYING` | Active gameplay with turns | Until tiles exhausted |
| `SCORING` | Final scoring calculation | Instant |
| `FINISHED` | Game completed | Persistent |

### Turn Timing

- **No Time Limits**: Players can take as long as needed
- **Bot Turns**: Processed automatically every 2 seconds
- **Disconnection Handling**: Turn skipped if player disconnected

## State Synchronization

### Synchronization Strategy

- **Full State**: Sent on room join and game start
- **Incremental Updates**: Sent for each game action
- **Conflict Resolution**: Server state is authoritative
- **Consistency**: All players receive identical state

### State Components

| Component | Sync Frequency | Recipients |
|-----------|----------------|------------|
| Room State | On change | All room members |
| Game State | Each turn | All game participants |
| Player State | On change | Specific player |
| Board State | Each placement | All game participants |

## Error Handling

### Error Categories

| Category | Code Range | Description |
|----------|------------|-------------|
| Connection | 1000-1099 | WebSocket and authentication errors |
| Room | 1100-1199 | Room management errors |
| Game | 1200-1299 | Game logic and rule violations |
| System | 1300-1399 | Server and system errors |

### Error Response Format

```json
{
  "type": "ERROR",
  "data": {
    "code": "INVALID_MOVE",
    "message": "Tile cannot be placed at this position",
    "details": {
      "position": {"x": 1, "y": 1},
      "reason": "Edge mismatch"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "messageId": "error-001"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `NOT_CONNECTED` | Player not authenticated |
| `ROOM_NOT_FOUND` | Invalid room ID |
| `ROOM_FULL` | Room at capacity |
| `GAME_ALREADY_STARTED` | Cannot join active game |
| `NOT_YOUR_TURN` | Action attempted out of turn |
| `INVALID_PLACEMENT` | Tile placement violates rules |
| `NO_MEEPLES` | Player has no available meeples |

## Data Types

### Position
```json
{
  "x": 0,
  "y": 0
}
```

### Tile
```json
{
  "id": 1,
  "north": "ROAD",
  "east": "FIELD", 
  "south": "CITY",
  "west": "FIELD",
  "features": [
    {
      "type": "ROAD",
      "edges": ["NORTH"],
      "id": 0
    }
  ],
  "hasMonastery": false,
  "hasShield": false
}
```

### Player
```json
{
  "id": "player-123",
  "name": "John",
  "color": "red",
  "meeples": 7,
  "score": 15,
  "isBot": false
}
```

### Game State
```json
{
  "tiles": {
    "0,0": {
      "tile": { /* Tile object */ },
      "position": {"x": 0, "y": 0},
      "rotation": 0,
      "meeples": []
    }
  },
  "currentTile": { /* Tile object */ },
  "players": [ /* Player objects */ ],
  "currentPlayer": 0,
  "gameStarted": true,
  "gameEnded": false,
  "scores": {
    "player-123": 15
  },
  "tilesLeft": 65
}
```

## Message Reference

### CONNECT
**Direction**: Client → Server  
**Purpose**: Authenticate and establish session

```json
{
  "type": "CONNECT",
  "data": {
    "playerId": "string",
    "name": "string", 
    "color": "string"
  }
}
```

### LIST_ROOMS
**Direction**: Client → Server  
**Purpose**: Request list of available rooms

```json
{
  "type": "LIST_ROOMS",
  "data": {}
}
```

**Response**:
```json
{
  "type": "LIST_ROOMS",
  "data": {
    "rooms": [
      {
        "id": "room-123",
        "name": "My Game",
        "playerCount": 2,
        "maxPlayers": 4,
        "gameStarted": false,
        "createdBy": "player-123"
      }
    ]
  }
}
```

### CREATE_ROOM
**Direction**: Client → Server  
**Purpose**: Create new game room

```json
{
  "type": "CREATE_ROOM",
  "data": {
    "roomName": "string",
    "maxPlayers": 4
  }
}
```

### JOIN_ROOM
**Direction**: Client → Server  
**Purpose**: Join existing room

```json
{
  "type": "JOIN_ROOM",
  "data": {
    "roomId": "string"
  }
}
```

### LEAVE_ROOM
**Direction**: Client → Server  
**Purpose**: Leave current room

```json
{
  "type": "LEAVE_ROOM",
  "data": {
    "roomId": "string"
  }
}
```

### ADD_BOT
**Direction**: Client → Server  
**Purpose**: Add AI player to room (host only)

```json
{
  "type": "ADD_BOT",
  "data": {
    "botName": "string",
    "difficulty": "easy|medium|hard"
  }
}
```

### GAME_START
**Direction**: Server → Client  
**Purpose**: Notify game has started

```json
{
  "type": "GAME_START",
  "data": {
    "roomId": "string",
    "players": [ /* Player objects */ ]
  }
}
```

### TURN_START
**Direction**: Server → Client  
**Purpose**: Begin new turn

```json
{
  "type": "TURN_START",
  "data": {
    "currentPlayer": "string",
    "currentTile": { /* Tile object */ },
    "validPlacements": [
      {
        "position": {"x": 1, "y": 0},
        "rotation": 0
      }
    ]
  }
}
```

### PLACE_TILE
**Direction**: Client → Server  
**Purpose**: Place tile on board

```json
{
  "type": "PLACE_TILE",
  "data": {
    "position": {"x": 1, "y": 0},
    "rotation": 90
  }
}
```

### PLACE_MEEPLE
**Direction**: Client → Server  
**Purpose**: Place meeple on tile

```json
{
  "type": "PLACE_MEEPLE",
  "data": {
    "featureId": 0
  }
}
```

### TURN_END
**Direction**: Server → Client  
**Purpose**: Turn completed

```json
{
  "type": "TURN_END",
  "data": {
    "playerId": "string",
    "scoreChange": 3,
    "nextPlayer": "string",
    "gameState": { /* GameState object */ }
  }
}
```

### GAME_END
**Direction**: Server → Client  
**Purpose**: Game completed

```json
{
  "type": "GAME_END",
  "data": {
    "winner": "string",
    "finalScore": {
      "player-123": 87,
      "player-456": 92
    },
    "gameState": { /* Final GameState */ }
  }
}
```

### ROOM_STATE
**Direction**: Server → Client  
**Purpose**: Room status update

```json
{
  "type": "ROOM_STATE",
  "data": {
    "roomId": "string",
    "players": [ /* Player objects */ ],
    "gameStarted": false,
    "gameEnded": false
  }
}
```

### GAME_STATE
**Direction**: Server → Client  
**Purpose**: Complete game state

```json
{
  "type": "GAME_STATE",
  "data": {
    "gameState": { /* GameState object */ }
  }
}
```

### PLAYER_UPDATE
**Direction**: Server → Client  
**Purpose**: Player-specific updates

```json
{
  "type": "PLAYER_UPDATE",
  "data": {
    "player": { /* Player object */ }
  }
}
```

## Game Rules Implementation

### Tile Placement Rules

1. **First Tile**: Starting tile placed at (0,0)
2. **Adjacency**: New tiles must be adjacent to existing tiles
3. **Edge Matching**: Adjacent edges must match (road-to-road, city-to-city, field-to-field)
4. **Rotation**: Tiles can be rotated in 90° increments
5. **Validation**: Server validates all placements

### Meeple Placement Rules

1. **Timing**: Only on the tile just placed
2. **Feature Availability**: Feature must not be connected to another meeple
3. **Meeple Limit**: Players have 7 meeples maximum
4. **Optional**: Meeple placement is optional

### Scoring Rules

#### Immediate Scoring (during game)
- **Completed Roads**: 1 point per tile
- **Completed Cities**: 2 points per tile (4 with shield)
- **Completed Monasteries**: 9 points (1 + 8 surrounding)

#### Final Scoring (game end)
- **Incomplete Roads**: 1 point per tile
- **Incomplete Cities**: 1 point per tile (2 with shield)
- **Incomplete Monasteries**: 1 point + 1 per surrounding tile
- **Fields**: 3 points per completed city supplied

### Bot AI Behavior

#### Easy Difficulty
- Random valid tile placement
- 50% chance to place meeple randomly

#### Medium Difficulty  
- Prefers monastery completion
- Strategic meeple placement
- 30% meeple placement rate

#### Hard Difficulty
- Advanced scoring optimization
- Feature completion priority
- Intelligent meeple management

## Examples

### Complete Game Session

```javascript
// 1. Connect to server
ws.send(JSON.stringify({
  type: "CONNECT",
  data: {
    playerId: "player-123",
    name: "Alice",
    color: "red"
  },
  timestamp: "2024-01-01T10:00:00Z",
  messageId: "msg-001"
}));

// 2. Create room
ws.send(JSON.stringify({
  type: "CREATE_ROOM", 
  data: {
    roomName: "Alice's Game",
    maxPlayers: 3
  },
  timestamp: "2024-01-01T10:00:01Z",
  messageId: "msg-002"
}));

// 3. Add bot
ws.send(JSON.stringify({
  type: "ADD_BOT",
  data: {
    botName: "Bot Alice",
    difficulty: "medium"
  },
  timestamp: "2024-01-01T10:00:02Z", 
  messageId: "msg-003"
}));

// 4. Place tile (during game)
ws.send(JSON.stringify({
  type: "PLACE_TILE",
  data: {
    position: {x: 1, y: 0},
    rotation: 90
  },
  timestamp: "2024-01-01T10:05:00Z",
  messageId: "msg-010"
}));

// 5. Place meeple
ws.send(JSON.stringify({
  type: "PLACE_MEEPLE",
  data: {
    featureId: 0
  },
  timestamp: "2024-01-01T10:05:01Z",
  messageId: "msg-011"
}));
```

### Error Handling Example

```javascript
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  if (message.type === "ERROR") {
    switch (message.data.code) {
      case "INVALID_PLACEMENT":
        console.error("Invalid tile placement:", message.data.message);
        // Show valid placements to user
        break;
      case "NOT_YOUR_TURN":
        console.warn("Wait for your turn");
        break;
      case "ROOM_FULL":
        console.error("Cannot join room - it's full");
        break;
    }
  }
};
```

## Protocol Compliance

### Client Implementation Requirements

1. **Message Format**: Must follow exact JSON schema
2. **Message IDs**: Must be unique per client session
3. **Timestamps**: Must be valid ISO 8601 format
4. **Error Handling**: Must handle all error codes gracefully
5. **State Management**: Must maintain local state consistency

### Server Guarantees

1. **Message Ordering**: Messages delivered in order per client
2. **State Consistency**: All clients receive identical game state
3. **Atomicity**: Game actions are atomic (all-or-nothing)
4. **Validation**: All game rules enforced server-side
5. **Cleanup**: Automatic cleanup of inactive sessions

---

**Protocol Version**: 1.0  
**Last Updated**: 2025-01-01  
**Compatibility**: WebSocket RFC 6455, JSON RFC 7159
