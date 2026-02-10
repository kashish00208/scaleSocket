package main

import (
	"sync"
	"time"
)

// Room represents a chat room
type Room struct {
	name         string
	clients      map[string]*Client
	mu           sync.RWMutex
	broadcast    chan *Message
	join         chan *Client
	leave        chan *Client
	messageCount int64
	createdAt    time.Time
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	room := &Room{
		name:      name,
		clients:   make(map[string]*Client),
		broadcast: make(chan *Message, 256),
		join:      make(chan *Client, 64),
		leave:     make(chan *Client, 64),
		createdAt: time.Now().UTC(),
	}

	go room.run()
	return room
}

// run handles the room's main event loop
func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			r.handleJoin(client)
		case client := <-r.leave:
			r.handleLeave(client)
		case msg := <-r.broadcast:
			r.handleBroadcast(msg)
		}
	}
}

// handleJoin adds a client to the room
func (r *Room) handleJoin(client *Client) {
	r.mu.Lock()
	r.clients[client.GetID()] = client
	r.mu.Unlock()

	// Notify others that user joined
	joinMsg := NewMessage(msgJoin, r.name, client.GetID(), client.GetUsername(), "joined the room")
	r.broadcast <- joinMsg

	// Send user list to all clients
	r.sendUserList()
}

// handleLeave removes a client from the room
func (r *Room) handleLeave(client *Client) {
	r.mu.Lock()
	if _, exists := r.clients[client.GetID()]; !exists {
		r.mu.Unlock()
		return
	}
	delete(r.clients, client.GetID())
	r.mu.Unlock()

	// Notify others that user left
	leaveMsg := NewMessage(msgLeave, r.name, client.GetID(), client.GetUsername(), "left the room")
	r.broadcast <- leaveMsg

	// Send updated user list
	r.sendUserList()
}

// handleBroadcast sends a message to all clients in the room
func (r *Room) handleBroadcast(msg *Message) {
	r.messageCount++
	r.mu.RLock()
	for _, client := range r.clients {
		client.SendMessage(msg)
	}
	r.mu.RUnlock()
}

// sendUserList sends the current user list to all clients
func (r *Room) sendUserList() {
	r.mu.RLock()
	users := make([]UserInfo, 0, len(r.clients))
	for _, client := range r.clients {
		users = append(users, UserInfo{
			UserId:   client.GetID(),
			UserName: client.GetUsername(),
			JoinedAT: time.Now().UTC(),
		})
	}
	r.mu.RUnlock()

	// Create user list message
	userListMsg := NewMessage(msgUserList, r.name, "", "", "")
	userListMsg.MetaData = map[string]interface{}{
		"users": users,
	}

	r.mu.RLock()
	for _, client := range r.clients {
		client.SendMessage(userListMsg)
	}
	r.mu.RUnlock()
}

// Broadcast sends a message to all clients in the room
func (r *Room) Broadcast(msg *Message) {
	select {
	case r.broadcast <- msg:
	default:
		// Channel full here
	}
}

// Join adds a client to the room
func (r *Room) Join(client *Client) {
	select {
	case r.join <- client:
	default:
		// Channel full
	}
}

// Leave removes a client from the room
func (r *Room) Leave(client *Client) {
	select {
	case r.leave <- client:
	default:
		// Channel full
	}
}

// GetUserCount returns the number of connected users
func (r *Room) GetUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// GetUsers returns a list of users in the room
func (r *Room) GetUsers() []UserInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]UserInfo, 0, len(r.clients))
	for _, client := range r.clients {
		users = append(users, UserInfo{
			UserId:   client.GetID(),
			UserName: client.GetUsername(),
			JoinedAT: time.Now().UTC(),
		})
	}
	return users
}

// GetStats returns statistics for the room
func (r *Room) GetStats() RoomStats {
	return RoomStats{
		Name:         r.name,
		UserCount:    r.GetUserCount(),
		MessageCount: r.messageCount,
		CreatedAt:    r.createdAt,
		Users:        r.GetUsers(),
	}
}

// IsEmpty returns true if the room has no clients
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients) == 0
}

// Close closes the room
func (r *Room) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.broadcast)
	close(r.join)
	close(r.leave)
}
