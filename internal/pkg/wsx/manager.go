package wsx

import (
	"sync"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client.UserID] = client
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.Unlock()
		}
	}
}

// SendDirect sends a message to a specific user.
func (h *Hub) SendDirect(senderID, receiverID string, message []byte) {
	h.RLock()
	target, ok := h.clients[receiverID]
	h.RUnlock()

	if ok {
		// Only forward if the target exists
		select {
		case target.Send <- message:
		default:
			// Buffer full, close connection to free resources
			h.Unregister(target)
		}
	}
	// If not online, message is dropped (as per requirement for now)
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}
