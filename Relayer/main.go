type Hub struct {
	// Ragistered client
	clients map[*Client]bool
	//Inbound messages
	broadcast chan []byte
	//Redis clinet
	redisClient *redis.client
}

func (h *Hub) Run() {
	pubsub := h.redisClient.Subsribe(c)
} 
