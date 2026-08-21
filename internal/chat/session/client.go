package session

import (
	"nurture/internal/chat/constant"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	Send           chan []byte
	UserID         string
	Channel        string
	delivered      map[string]struct{}
	deliveredOrder []string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, channel string) *Client {
	return &Client{
		Hub:     hub,
		Conn:    conn,
		Send:    make(chan []byte, constant.WSSendBufferSize),
		UserID:  userID,
		Channel: channel,
	}
}

func (c *Client) TrySend(message []byte) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	select {
	case c.Send <- message:
		return true
	default:
		return false
	}
}

func (c *Client) markDelivered(eventID string) bool {
	if eventID == "" {
		return true
	}
	if c.delivered == nil {
		c.delivered = make(map[string]struct{})
	}
	if _, ok := c.delivered[eventID]; ok {
		return false
	}
	c.delivered[eventID] = struct{}{}
	c.deliveredOrder = append(c.deliveredOrder, eventID)
	if len(c.deliveredOrder) > constant.WSDeliveredCache {
		oldest := c.deliveredOrder[0]
		delete(c.delivered, oldest)
		copy(c.deliveredOrder, c.deliveredOrder[1:])
		c.deliveredOrder = c.deliveredOrder[:len(c.deliveredOrder)-1]
	}
	return true
}

func (c *Client) ReadPump(handleMessage func(message []byte)) {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(constant.WSMaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(constant.WSPongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(constant.WSPongWait))
		return nil
	})
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		if handleMessage != nil {
			handleMessage(message)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(constant.WSPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(constant.WSWriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(constant.WSWriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
