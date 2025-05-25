# Carcassonne WebSocket Server

A multiplayer Carcassonne game server implemented in Go with WebSocket support, featuring real-time gameplay, bot AI, and Docker deployment.

## Features

- **Real-time Multiplayer**: WebSocket-based communication for instant game updates
- **Room Management**: Create, join, and manage game rooms with 2-5 players
- **Bot AI**: Add AI players with configurable difficulty levels
- **Full Game Logic**: Complete Carcassonne tile placement and scoring system
- **Docker Ready**: Containerized for easy deployment on AWS ECS or any container platform
- **RESTful API**: HTTP endpoints for health checks and room listing
- **Web Client**: Built-in HTML/JavaScript client for testing

## Architecture

```
carcassonne-ws/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── api/            # HTTP handlers and routing
│   ├── game/           # Core game logic (tiles, board, scoring)
│   ├── player/         # Player and bot management
│   ├── room/           # Room and game session management
│   └── websocket/      # WebSocket hub and client handling
├── web/static/         # Static web client files
├── Dockerfile          # Container configuration
├── docker-compose.yml  # Development environment
└── go.mod             # Go module dependencies
```

## Quick Start

### Local Development

1. **Clone and build**:
```bash
git clone <repository>
cd carcassonne-ws
go mod tidy
go run cmd/server/main.go
```

2. **Access the application**:
- WebSocket endpoint: `ws://localhost:8080/ws`
- Web client: `http://localhost:8080`
- Health check: `http://localhost:8080/health`

### Docker Development

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build and run manually
docker build -t carcassonne-ws .
docker run -p 8080:8080 carcassonne-ws
```

### AWS ECS Deployment

1. **Build and push to ECR**:
```bash
# Create ECR repository
aws ecr create-repository --repository-name carcassonne-ws

# Get login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build and tag
docker build -t carcassonne-ws .
docker tag carcassonne-ws:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/carcassonne-ws:latest

# Push
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/carcassonne-ws:latest
```

2. **Create ECS Task Definition**:
```json
{
  "family": "carcassonne-ws",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::<account-id>:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "carcassonne-ws",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/carcassonne-ws:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "PORT",
          "value": "8080"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/carcassonne-ws",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

## WebSocket API

### Message Format
```json
{
  "type": "MESSAGE_TYPE",
  "data": { /* message-specific payload */ },
  "timestamp": "2024-01-01T00:00:00Z",
  "messageId": "unique-id"
}
```

### Message Types

#### Connection & Room Management
- `CONNECT` - Initial connection with player info
- `LIST_ROOMS` - Request active rooms
- `CREATE_ROOM` - Create new game room
- `JOIN_ROOM` - Join existing room
- `LEAVE_ROOM` - Leave current room
- `ADD_BOT` - Add AI player (room creator only)

#### Game Flow
- `GAME_START` - Game begins notification
- `TURN_START` - New turn with tile data
- `PLACE_TILE` - Player tile placement
- `PLACE_MEEPLE` - Player meeple placement
- `TURN_END` - Turn completion
- `GAME_END` - Game completion

#### State Synchronization
- `ROOM_STATE` - Current room status
- `GAME_STATE` - Current board state
- `PLAYER_UPDATE` - Player-specific updates

### Example Usage

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

// Send connection message
ws.send(JSON.stringify({
  type: 'CONNECT',
  data: {
    playerId: 'player123',
    name: 'John',
    color: 'red'
  },
  timestamp: new Date().toISOString(),
  messageId: 'msg123'
}));

// Create a room
ws.send(JSON.stringify({
  type: 'CREATE_ROOM',
  data: {
    roomName: 'My Game',
    maxPlayers: 4
  },
  timestamp: new Date().toISOString(),
  messageId: 'msg124'
}));
```

## Game Rules Implementation

### Tile System
- 72 unique tiles with roads, cities, monasteries, and fields
- Tile rotation and placement validation
- Edge matching requirements

### Scoring System
- **Roads**: 1 point per tile when completed
- **Cities**: 2 points per tile when completed (4 points with shield)
- **Monasteries**: 1 point per surrounding tile (max 9)
- **Fields**: Points based on completed cities they supply

### Bot AI
- **Easy**: Random valid moves
- **Medium**: Prioritizes completing features
- **Hard**: Advanced scoring optimization (extensible)

## API Endpoints

- `GET /health` - Health check
- `GET /api/rooms` - List active rooms (HTTP fallback)
- `WS /ws` - WebSocket connection

## Configuration

Environment variables:
- `PORT` - Server port (default: 8080)

## Development

### Running Tests
```bash
go test ./...
```

### Adding New Features
1. Game logic goes in `internal/game/`
2. WebSocket messages in `internal/websocket/messages.go`
3. Room management in `internal/room/`
4. Bot AI in `internal/player/bot.go`

### Code Structure
- **Separation of Concerns**: Game logic, networking, and room management are separate
- **Concurrent Safe**: All shared state uses proper synchronization
- **Extensible**: Easy to add new tile types, scoring rules, or bot strategies

## Performance Considerations

- **Memory Usage**: Games are stored in memory (no persistence)
- **Concurrency**: Supports multiple concurrent games
- **Scalability**: Single instance design (can be extended with Redis for multi-instance)

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure server is running on correct port
2. **WebSocket upgrade failed**: Check CORS settings and protocol
3. **Bot moves not processing**: Check bot ticker interval in hub.go

### Logs
The server logs all connections, disconnections, and game events to stdout.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

MIT License - see LICENSE file for details.
