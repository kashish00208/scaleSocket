package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Room struct {
	Sockets []*websocket.Conn
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Error in Upgrading websocket server")
		return
	}

	defer conn.Close()

	go handleConnection(conn)

}

func handleConnection(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			fmt.Println("Error reading messae", err)
			break
		}

		fmt.Printf("Received %s\\n", msg)
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			println("Error writing messages : ", err)
			break
		}
	}
}
func main() {
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("Websocket server started on server 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Print("error starting the server", err)
	}
}
