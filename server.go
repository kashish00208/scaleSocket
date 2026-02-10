package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ServerStats holds statistics for the server
type ServerStats struct {
	TotalConnections int64                `json:"total_connections"`
	ActiveRooms      int                  `json:"active_rooms"`
	TotalMessages    int64                `json:"total_messages"`
	Rooms            map[string]RoomStats `json:"rooms"`
	UpSince          time.Time            `json:"up_since"`
	Uptime           string               `json:"uptime"`
}

// ChatServer manages all rooms and client connections
type ChatServer struct {
	rooms            map[string]*Room
	mu               sync.RWMutex
	clients          map[string]*Client
	clientsMu        sync.RWMutex
	totalMessages    int64
	totalConnections int64
	startTime        time.Time
	upgrader         websocket.Upgrader
}

// NewChatServer creates a new chat server
func NewChatServer() *ChatServer {
	return &ChatServer{
		rooms:     make(map[string]*Room),
		clients:   make(map[string]*Client),
		startTime: time.Now().UTC(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// HandleWebSocket handles a new WebSocket connection
func (s *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := NewClient(conn, s)

	s.clientsMu.Lock()
	s.clients[client.GetID()] = client
	s.clientsMu.Unlock()

	atomic.AddInt64(&s.totalConnections, 1)

	log.Printf("Client %s connected. Total connections: %d", client.GetID(), atomic.LoadInt64(&s.totalConnections))

	// Start read and write pumps
	go client.WritePump()
	client.ReadPump()

	// Clean up on disconnect
	s.clientsMu.Lock()
	delete(s.clients, client.GetID())
	s.clientsMu.Unlock()
}

// JoinRoom adds a client to a room
func (s *ChatServer) JoinRoom(client *Client) {
	roomName := client.GetRoom()
	if roomName == "" {
		client.sendError("Room name required")
		return
	}

	s.mu.Lock()
	room, exists := s.rooms[roomName]
	if !exists {
		room = NewRoom(roomName)
		s.rooms[roomName] = room
	}
	s.mu.Unlock()

	room.Join(client)
	log.Printf("Client %s joined room %s", client.GetID(), roomName)
}

// LeaveRoom removes a client from a room
func (s *ChatServer) LeaveRoom(client *Client) {
	roomName := client.GetRoom()
	if roomName == "" {
		return
	}

	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.Leave(client)

	// Clean up empty rooms
	if room.IsEmpty() {
		s.mu.Lock()
		delete(s.rooms, roomName)
		s.mu.Unlock()
		room.Close()
		log.Printf("Room %s closed (empty)", roomName)
	}

	log.Printf("Client %s left room %s", client.GetID(), roomName)
}

// BroadcastToRoom sends a message to all clients in a room
func (s *ChatServer) BroadcastToRoom(roomName string, msg *Message) {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return
	}

	atomic.AddInt64(&s.totalMessages, 1)
	room.Broadcast(msg)
}

// GetRoom returns a room by name
func (s *ChatServer) GetRoom(name string) *Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[name]
}

// RoomExists checks if a room exists
func (s *ChatServer) RoomExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.rooms[name]
	return exists
}

// ListRooms returns a list of all rooms
func (s *ChatServer) ListRooms() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]string, 0, len(s.rooms))
	for name := range s.rooms {
		rooms = append(rooms, name)
	}
	return rooms
}

// ConnectionCount returns the current number of connected clients
func (s *ChatServer) ConnectionCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// RoomCount returns the current number of rooms
func (s *ChatServer) RoomCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.rooms)
}

// GetStats returns server statistics
func (s *ChatServer) GetStats() string {
	s.mu.RLock()
	rooms := make(map[string]RoomStats)
	for name, room := range s.rooms {
		rooms[name] = room.GetStats()
	}
	s.mu.RUnlock()

	stats := ServerStats{
		TotalConnections: atomic.LoadInt64(&s.totalConnections),
		ActiveRooms:      len(rooms),
		TotalMessages:    atomic.LoadInt64(&s.totalMessages),
		Rooms:            rooms,
		UpSince:          s.startTime,
		Uptime:           time.Since(s.startTime).String(),
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	return string(data)
}

// Close gracefully shuts down the server
func (s *ChatServer) Close() {
	s.mu.Lock()
	for _, room := range s.rooms {
		room.Close()
	}
	s.rooms = make(map[string]*Room)
	s.mu.Unlock()

	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()
}

// GetRoomUsers returns all users in a specific room
func (s *ChatServer) GetRoomUsers(roomName string) []UserInfo {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return []UserInfo{}
	}

	return room.GetUsers()
}

// KickClientFromRoom removes a specific client from a room
func (s *ChatServer) KickClientFromRoom(clientID, roomName string) bool {
	s.clientsMu.RLock()
	client, exists := s.clients[clientID]
	s.clientsMu.RUnlock()

	if !exists {
		return false
	}

	if client.GetRoom() == roomName {
		s.LeaveRoom(client)
		return true
	}

	return false
}
