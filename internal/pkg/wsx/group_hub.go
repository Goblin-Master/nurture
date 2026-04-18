package wsx

import (
	"sync"
)

type groupSubscription struct {
	client  *GroupClient
	groupID string
}

type groupBroadcast struct {
	groupID string
	data    []byte
}

type GroupHub struct {
	mu sync.RWMutex

	clients map[string]*GroupClient
	groups  map[string]map[*GroupClient]struct{}

	register   chan *GroupClient
	unregister chan *GroupClient
	subscribe  chan groupSubscription
	unsubscribe chan groupSubscription
	broadcast  chan groupBroadcast
}

func NewGroupHub() *GroupHub {
	return &GroupHub{
		clients:    make(map[string]*GroupClient),
		groups:     make(map[string]map[*GroupClient]struct{}),
		register:   make(chan *GroupClient),
		unregister: make(chan *GroupClient),
		subscribe:  make(chan groupSubscription),
		unsubscribe: make(chan groupSubscription),
		broadcast:  make(chan groupBroadcast),
	}
}

func (h *GroupHub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if old, ok := h.clients[c.UserID]; ok && old != c {
				old.Close()
			}
			h.clients[c.UserID] = c
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if cur, ok := h.clients[c.UserID]; ok && cur == c {
				delete(h.clients, c.UserID)
			}
			for gid, set := range h.groups {
				if _, ok := set[c]; ok {
					delete(set, c)
					if len(set) == 0 {
						delete(h.groups, gid)
					}
				}
			}
			h.mu.Unlock()
		case s := <-h.subscribe:
			h.mu.Lock()
			set, ok := h.groups[s.groupID]
			if !ok {
				set = make(map[*GroupClient]struct{})
				h.groups[s.groupID] = set
			}
			set[s.client] = struct{}{}
			h.mu.Unlock()
		case s := <-h.unsubscribe:
			h.mu.Lock()
			if set, ok := h.groups[s.groupID]; ok {
				delete(set, s.client)
				if len(set) == 0 {
					delete(h.groups, s.groupID)
				}
			}
			h.mu.Unlock()
		case b := <-h.broadcast:
			h.mu.RLock()
			set := h.groups[b.groupID]
			for c := range set {
				select {
				case c.Send <- b.data:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *GroupHub) RegisterClient(c *GroupClient) {
	h.register <- c
}

func (h *GroupHub) UnregisterClient(c *GroupClient) {
	h.unregister <- c
}

func (h *GroupHub) Subscribe(c *GroupClient, groupID string) {
	h.subscribe <- groupSubscription{client: c, groupID: groupID}
}

func (h *GroupHub) Unsubscribe(c *GroupClient, groupID string) {
	h.unsubscribe <- groupSubscription{client: c, groupID: groupID}
}

func (h *GroupHub) Broadcast(groupID string, data []byte) {
	h.broadcast <- groupBroadcast{groupID: groupID, data: data}
}

