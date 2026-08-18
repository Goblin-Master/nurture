package chat

type subscription struct {
	session *session
	groupID string
}

type userMessage struct {
	channel string
	userID  string
	data    []byte
}

type groupMessage struct {
	groupID string
	data    []byte
}

type sessionHub struct {
	clients map[string]map[string]map[*session]struct{}
	groups  map[string]map[*session]struct{}

	register    chan *session
	unregister  chan *session
	subscribe   chan subscription
	unsubscribe chan subscription
	sendUser    chan userMessage
	broadcast   chan groupMessage
}

func newSessionHub() *sessionHub {
	return &sessionHub{
		clients:     make(map[string]map[string]map[*session]struct{}),
		groups:      make(map[string]map[*session]struct{}),
		register:    make(chan *session),
		unregister:  make(chan *session),
		subscribe:   make(chan subscription),
		unsubscribe: make(chan subscription),
		sendUser:    make(chan userMessage),
		broadcast:   make(chan groupMessage),
	}
}

func (h *sessionHub) run() {
	for {
		select {
		case client := <-h.register:
			h.addSession(client)
		case client := <-h.unregister:
			h.removeSession(client)
		case sub := <-h.subscribe:
			h.addSubscription(sub.session, sub.groupID)
		case sub := <-h.unsubscribe:
			h.removeSubscription(sub.session, sub.groupID)
		case msg := <-h.sendUser:
			h.sendToUser(msg.channel, msg.userID, msg.data)
		case msg := <-h.broadcast:
			h.broadcastToGroup(msg.groupID, msg.data)
		}
	}
}

func (h *sessionHub) registerSession(client *session) {
	h.register <- client
}

func (h *sessionHub) unregisterSession(client *session) {
	h.unregister <- client
}

func (h *sessionHub) subscribeGroup(client *session, groupID string) {
	h.subscribe <- subscription{session: client, groupID: groupID}
}

func (h *sessionHub) unsubscribeGroup(client *session, groupID string) {
	h.unsubscribe <- subscription{session: client, groupID: groupID}
}

func (h *sessionHub) sendToUserChannel(channel, userID string, data []byte) {
	h.sendUser <- userMessage{channel: channel, userID: userID, data: data}
}

func (h *sessionHub) broadcastGroup(groupID string, data []byte) {
	h.broadcast <- groupMessage{groupID: groupID, data: data}
}

func (h *sessionHub) addSession(client *session) {
	byChannel, ok := h.clients[client.channel]
	if !ok {
		byChannel = make(map[string]map[*session]struct{})
		h.clients[client.channel] = byChannel
	}
	byUser, ok := byChannel[client.userID]
	if !ok {
		byUser = make(map[*session]struct{})
		byChannel[client.userID] = byUser
	}
	byUser[client] = struct{}{}
}

func (h *sessionHub) removeSession(client *session) {
	if byChannel, ok := h.clients[client.channel]; ok {
		if byUser, ok := byChannel[client.userID]; ok {
			if _, ok := byUser[client]; ok {
				delete(byUser, client)
				close(client.send)
			}
			if len(byUser) == 0 {
				delete(byChannel, client.userID)
			}
		}
		if len(byChannel) == 0 {
			delete(h.clients, client.channel)
		}
	}
	for groupID, clients := range h.groups {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.groups, groupID)
			}
		}
	}
}

func (h *sessionHub) addSubscription(client *session, groupID string) {
	if groupID == "" {
		return
	}
	clients, ok := h.groups[groupID]
	if !ok {
		clients = make(map[*session]struct{})
		h.groups[groupID] = clients
	}
	clients[client] = struct{}{}
}

func (h *sessionHub) removeSubscription(client *session, groupID string) {
	if clients, ok := h.groups[groupID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.groups, groupID)
		}
	}
}

func (h *sessionHub) sendToUser(channel, userID string, data []byte) {
	byChannel, ok := h.clients[channel]
	if !ok {
		return
	}
	byUser, ok := byChannel[userID]
	if !ok {
		return
	}
	for client := range byUser {
		h.send(client, data)
	}
}

func (h *sessionHub) broadcastToGroup(groupID string, data []byte) {
	clients, ok := h.groups[groupID]
	if !ok {
		return
	}
	for client := range clients {
		h.send(client, data)
	}
}

func (h *sessionHub) send(client *session, data []byte) {
	if !client.trySend(data) {
		h.removeSession(client)
		if client.conn != nil {
			_ = client.conn.Close()
		}
	}
}
