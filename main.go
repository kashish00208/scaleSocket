package main

import (
	"fmt"

	"golang.org/x/net/websocket"
)

type Server struct {
	conn map[*websocket.Conn]bool
}

func newServer() *Server {
	return &Server{
		conn: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handler(ws *websocket.Conn) {
	fmt.Printf("")
}

func main() {

}
