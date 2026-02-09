package main
import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"github.com/gorilla/websocket"
)

type ChatServerStruct struct {
	rooms  map[string]*Room
	
}