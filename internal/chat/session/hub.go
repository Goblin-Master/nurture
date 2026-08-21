package session

type subscription struct {
	client *Client
	roomID string
}

type userMessage struct {
	channel string
	userID  string
	eventID string
	data    []byte
}

type roomMessage struct {
	roomID  string
	eventID string
	data    []byte
}

type Hub struct {
	clients map[string]map[string]map[*Client]struct{}
	rooms   map[string]map[*Client]struct{}

	register    chan *Client
	unregister  chan *Client
	subscribe   chan subscription
	unsubscribe chan subscription
	sendUser    chan userMessage
	broadcast   chan roomMessage
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[string]map[*Client]struct{}),
		rooms:       make(map[string]map[*Client]struct{}),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		subscribe:   make(chan subscription),
		unsubscribe: make(chan subscription),
		sendUser:    make(chan userMessage),
		broadcast:   make(chan roomMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case sub := <-h.subscribe:
			h.addSubscription(sub.client, sub.roomID)
		case sub := <-h.unsubscribe:
			h.removeSubscription(sub.client, sub.roomID)
		case msg := <-h.sendUser:
			h.sendToUser(msg.channel, msg.userID, msg.eventID, msg.data)
		case msg := <-h.broadcast:
			h.broadcastToRoom(msg.roomID, msg.eventID, msg.data)
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Subscribe(client *Client, roomID string) {
	h.subscribe <- subscription{client: client, roomID: roomID}
}

func (h *Hub) Unsubscribe(client *Client, roomID string) {
	h.unsubscribe <- subscription{client: client, roomID: roomID}
}

func (h *Hub) SendToUser(channel, userID string, data []byte) {
	h.sendUser <- userMessage{channel: channel, userID: userID, data: data}
}

func (h *Hub) DeliverToUser(channel, userID, eventID string, data []byte) {
	h.sendUser <- userMessage{channel: channel, userID: userID, eventID: eventID, data: data}
}

func (h *Hub) Broadcast(roomID string, data []byte) {
	h.broadcast <- roomMessage{roomID: roomID, data: data}
}

func (h *Hub) DeliverToRoom(roomID, eventID string, data []byte) {
	h.broadcast <- roomMessage{roomID: roomID, eventID: eventID, data: data}
}

func (h *Hub) addClient(client *Client) {
	byChannel, ok := h.clients[client.Channel]
	if !ok {
		byChannel = make(map[string]map[*Client]struct{})
		h.clients[client.Channel] = byChannel
	}
	byUser, ok := byChannel[client.UserID]
	if !ok {
		byUser = make(map[*Client]struct{})
		byChannel[client.UserID] = byUser
	}
	byUser[client] = struct{}{}
}

func (h *Hub) removeClient(client *Client) {
	if byChannel, ok := h.clients[client.Channel]; ok {
		if byUser, ok := byChannel[client.UserID]; ok {
			if _, ok := byUser[client]; ok {
				delete(byUser, client)
				close(client.Send)
			}
			if len(byUser) == 0 {
				delete(byChannel, client.UserID)
			}
		}
		if len(byChannel) == 0 {
			delete(h.clients, client.Channel)
		}
	}
	for roomID, clients := range h.rooms {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.rooms, roomID)
			}
		}
	}
}

func (h *Hub) addSubscription(client *Client, roomID string) {
	if roomID == "" {
		return
	}
	clients, ok := h.rooms[roomID]
	if !ok {
		clients = make(map[*Client]struct{})
		h.rooms[roomID] = clients
	}
	clients[client] = struct{}{}
}

func (h *Hub) removeSubscription(client *Client, roomID string) {
	if clients, ok := h.rooms[roomID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

func (h *Hub) sendToUser(channel, userID, eventID string, data []byte) {
	byChannel, ok := h.clients[channel]
	if !ok {
		return
	}
	byUser, ok := byChannel[userID]
	if !ok {
		return
	}
	for client := range byUser {
		h.send(client, eventID, data)
	}
}

func (h *Hub) broadcastToRoom(roomID, eventID string, data []byte) {
	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}
	for client := range clients {
		h.send(client, eventID, data)
	}
}

func (h *Hub) send(client *Client, eventID string, data []byte) {
	if !client.markDelivered(eventID) {
		return
	}
	if !client.TrySend(data) {
		h.removeClient(client)
		if client.Conn != nil {
			_ = client.Conn.Close()
		}
	}
}
