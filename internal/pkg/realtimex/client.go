package realtimex

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	ChannelDirect = "direct"
	ChannelGroup  = "group"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufferSize = 256
)

type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	Send    chan []byte
	UserID  string
	Channel string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, channel string) *Client {
	return &Client{
		Hub:     hub,
		Conn:    conn,
		Send:    make(chan []byte, sendBufferSize),
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

func (c *Client) ReadPump(handleMessage func(message []byte)) {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
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
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
