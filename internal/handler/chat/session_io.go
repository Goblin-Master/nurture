package chat

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	directChannel = "direct"
	groupChannel  = "group"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufferSize = 256
)

type session struct {
	hub     *sessionHub
	conn    *websocket.Conn
	send    chan []byte
	userID  string
	channel string
}

func newSession(hub *sessionHub, conn *websocket.Conn, userID, channel string) *session {
	return &session{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, sendBufferSize),
		userID:  userID,
		channel: channel,
	}
}

func (c *session) trySend(message []byte) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	select {
	case c.send <- message:
		return true
	default:
		return false
	}
}

func (c *session) readMessages(handleMessage func(message []byte)) {
	defer func() {
		c.hub.unregisterSession(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if handleMessage != nil {
			handleMessage(message)
		}
	}
}

func (c *session) writeMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
