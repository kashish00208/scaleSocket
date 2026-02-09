package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	id       string
	username string
	room     string
	conn     *websocket.Conn
	send     chan *Message
	mu       sync.RWMutex
	closed   bool
	server   *ChatServer
}

func NewClient(conn *websocket.Conn, server *ChatServer) *Client {
	return &Client{
		id:     generateUserID(),
		conn:   conn,
		send:   make(chan *Message, 256),
		server: server,
		closed: false,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		// Update read deadline on successful read
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		switch msg.Type {
		case MsgTypeJoin:
			c.handleJoin(&msg)
		case MsgTypeLeave:
			c.handleLeave(&msg)
		case MsgTypeChat:
			c.handleChat(&msg)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// WritePump sends messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleJoin processes a user joining a room
func (c *Client) handleJoin(msg *Message) {
	c.mu.Lock()
	c.room = msg.Room
	c.username = msg.Username
	c.id = msg.UserID
	c.mu.Unlock()

	c.server.JoinRoom(c)
}

// handleLeave processes a user leaving a room
func (c *Client) handleLeave(msg *Message) {
	c.server.LeaveRoom(c)
}

// handleChat processes a chat message
func (c *Client) handleChat(msg *Message) {
	c.mu.RLock()
	room := c.room
	userID := c.id
	username := c.username
	c.mu.RUnlock()

	if room == "" {
		c.sendError("Not joined to any room")
		return
	}

	chatMsg := NewMessage(MsgTypeBroadcast, room, userID, username, msg.Content)
	chatMsg.Metadata = msg.Metadata

	c.server.BroadcastToRoom(room, chatMsg)
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg *Message) {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	if closed {
		return
	}

	select {
	case c.send <- msg:
	default:
		// Channel full, close the client
		log.Printf("Client %s send channel full, closing", c.id)
		c.Close()
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	msg := NewMessage(MsgTypeError, "", c.id, "", errMsg)
	c.SendMessage(msg)
}

// Close closes the client connection
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	room := c.room
	c.mu.Unlock()

	close(c.send)

	if room != "" {
		c.server.LeaveRoom(c)
	}
}

// GetID returns the client ID
func (c *Client) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

// GetUsername returns the client username
func (c *Client) GetUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.username
}

// GetRoom returns the current room
func (c *Client) GetRoom() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.room
}
