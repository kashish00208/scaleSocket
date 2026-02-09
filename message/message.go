package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	msgJoin      MessageType = "join"
	msgLeave     MessageType = "leave"
	msgChat      MessageType = "chat"
	msgBroadcast MessageType = "broadcast"
	msgError     MessageType = "Error"
	msgUserList  MessageType = "userList"
	msgAck       MessageType = "Ack"
)

type Message struct {
	Type      MessageType            `json:"type"`
	Room      string                 `json:"room"`
	UserId    string                 `json:"user_id"`
	UserName  string                 `json:"username"`
	Content   string                 `json:"content"`
	TimeStamp string                 `json:"timestamp"`
	MsgID     string                 `json:"message_id"`
	MetaData  map[string]interface{} `json:"metadata,omitempty"`
}

type ClientJoinMessage struct {
	Room     string                 `json:"room"`
	UserId   string                 `json:"user_id"`
	Content  string                 `json:"content"`
	MetaData map[string]interface{} `json:"metadata,omitempty"`
}

type UserInfo struct {
	UserId   string    `json:"user_id"`
	UserName string    `json:"username"`
	JoinedAT time.Time `json:"joined_at"`
}

type RoomStats struct {
	Name         string     `json:"name"`
	UserCount    int        `json:"user_count"`
	MessageCount int64      `json:"message_count"`
	CreatedAt    time.Time  `json:"created_at"`
	Users        []UserInfo `json:"users"`
}

type ServerStatus struct {
	TotalConnections int64                `json:"total_connections"`
	ActiveRooms      int                  `json:"active_rooms"`
	TotalMessages    int64                `json:"total_messages"`
	Rooms            map[string]RoomStats `json:"rooms"`
	UpSince          time.Time            `json:"up_since"`
	Uptime           string               `json:"uptime"`
}

func NewMessage(msgType MessageType, room, userID, username, content string) *Message {
	return &Message{
		Type:      msgType,
		Room:      room,
		UserId:    userID,
		Content:   content,
		TimeStamp: time.Now().UTC().Format(time.RFC3339),
		MsgID:     uuid.New().String(),
	}
}

func (m *Message) JSON() []byte {
	data, _ := json.Marshal(m)
	return data
}

func FromJSON(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
