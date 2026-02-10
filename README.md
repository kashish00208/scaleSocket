# Scalable Real-Time Chat Engine

A high-performance, production-ready chat engine built with Go and WebSockets. Designed for horizontal scalability, low latency, and high throughput.

## Features

### Core Features
- **WebSocket-based real-time communication**: Bidirectional, low-latency messaging
- **Multi-room support**: Organize users into isolated chat rooms
- **User management**: Track connected users and their presence
- **Graceful connection handling**: Proper cleanup and resource management
- **Thread-safe operations**: Concurrent safe with minimal locking
- **Connection pooling**: Buffered channels prevent blocking
- **Auto room cleanup**: Empty rooms are automatically removed

### Scalability Features
- **Non-blocking message broadcasting**: Goroutine-per-connection model
- **Efficient channel management**: Sized channels to prevent memory bloat
- **Atomic counters**: Lock-free statistics tracking
- **Connection limits handling**: Graceful degradation under high load
- **Heartbeat/ping-pong**: Keep-alive mechanism to detect stale connections

### Monitoring & Statistics
- **Health endpoint**: Real-time server health checks
- **Detailed stats endpoint**: Message counts, room stats, user info
- **Per-room statistics**: User counts, message counts, creation time
- **Connection tracking**: Total connections across all time

## Architecture

### Components

**Client (`client.go`)**: Represents a connected WebSocket client
- Manages individual connection lifecycle
- Implements read/write pumps for async I/O
- Handles message routing and buffering
- Supports graceful disconnection

**Room (`room.go`)**: Manages a group of connected clients
- Event-driven architecture with channels
- Efficient message broadcasting
- User presence tracking
- Automatic cleanup when empty

**Server (`server.go`)**: Central hub managing all rooms and connections
- Room lifecycle management
- Client connection handling
- Message statistics
- Health monitoring

**Message (`message.go`)**: Protocol definitions and data structures
- JSON serialization
- Message type definitions
- Metadata support

## Building

### Prerequisites
- Go 1.21 or later
- `gorilla/websocket` package
- `google/uuid` package


