package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func NewChatServer() {
	//will write the logic here
}

func main() {
	server := NewChatServer()

	//http handlers ragister

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"Status":"Healthy","connections":%d,"rooms":%d}`,
			server.ConnectionsCount(),
			server.RoomCount(),
		)
	})

	http.HandleFunc("status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		status := server.getStats()
		fmt.Print(w, status)
	})

	port := ":8080"

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	server.Close()
	log.Println("Server closed")

}
