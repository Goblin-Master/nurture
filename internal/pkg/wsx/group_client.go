package wsx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

type GroupClient struct {
	Hub    *GroupHub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string

	OnSubscribe func(ctx context.Context, groupID string) error
	OnSend      func(ctx context.Context, groupID, messageID, msgType, content string, now int64) error
}

func (c *GroupClient) Close() {
	_ = c.Conn.Close()
	close(c.Send)
}

type groupInMessage struct {
	Op       string   `json:"op"`
	GroupID  string   `json:"group_id"`
	GroupIDs []string `json:"group_ids"`
	MessageID string  `json:"message_id"`
	Type     string   `json:"type"`
	Content  string   `json:"content"`
}

type groupAckMessage struct {
	Op       string              `json:"op"`
	For      string              `json:"for"`
	Ok       bool                `json:"ok"`
	Error    string              `json:"error,omitempty"`
	GroupID  string              `json:"group_id,omitempty"`
	GroupIDs []string            `json:"group_ids,omitempty"`
	Failed   []groupFailedDetail `json:"failed,omitempty"`
	MessageID string             `json:"message_id,omitempty"`
	ServerTS int64               `json:"server_ts,omitempty"`
}

type groupFailedDetail struct {
	GroupID string `json:"group_id"`
	Error   string `json:"error"`
}

type groupOutMessage struct {
	Op      string           `json:"op"`
	GroupID string           `json:"group_id"`
	Message groupMessageBody `json:"message"`
}

type groupMessageBody struct {
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}

func (c *GroupClient) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		_ = c.Conn.Close()
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		var in groupInMessage
		if err := json.Unmarshal(msg, &in); err != nil {
			c.writeAck(groupAckMessage{Op: "ack", For: "parse", Ok: false, Error: "invalid_json"})
			continue
		}
		switch in.Op {
		case "subscribe":
			c.handleSubscribe(in.GroupID)
		case "subscribe_many":
			c.handleSubscribeMany(in.GroupIDs)
		case "unsubscribe":
			if in.GroupID != "" {
				c.Hub.Unsubscribe(c, in.GroupID)
			}
		case "send":
			c.handleSend(in.GroupID, in.MessageID, in.Type, in.Content)
		default:
			c.writeAck(groupAckMessage{Op: "ack", For: in.Op, Ok: false, Error: "unknown_op"})
		}
	}
}

func (c *GroupClient) WritePump() {
	defer func() {
		_ = c.Conn.Close()
	}()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *GroupClient) handleSubscribe(groupID string) {
	if groupID == "" {
		c.writeAck(groupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: "missing_group_id"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if c.OnSubscribe != nil {
		if err := c.OnSubscribe(ctx, groupID); err != nil {
			c.writeAck(groupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: err.Error(), GroupID: groupID})
			return
		}
	}
	c.Hub.Subscribe(c, groupID)
	c.writeAck(groupAckMessage{Op: "ack", For: "subscribe", Ok: true, GroupID: groupID})
}

func (c *GroupClient) handleSubscribeMany(groupIDs []string) {
	if len(groupIDs) == 0 {
		c.writeAck(groupAckMessage{Op: "ack", For: "subscribe_many", Ok: false, Error: "missing_group_ids"})
		return
	}
	okIDs := make([]string, 0, len(groupIDs))
	failed := make([]groupFailedDetail, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, gid := range groupIDs {
		if gid == "" {
			continue
		}
		if c.OnSubscribe != nil {
			if err := c.OnSubscribe(ctx, gid); err != nil {
				failed = append(failed, groupFailedDetail{GroupID: gid, Error: err.Error()})
				continue
			}
		}
		c.Hub.Subscribe(c, gid)
		okIDs = append(okIDs, gid)
	}
	c.writeAck(groupAckMessage{Op: "ack", For: "subscribe_many", Ok: len(failed) == 0, GroupIDs: okIDs, Failed: failed})
}

func (c *GroupClient) handleSend(groupID, messageID, msgType, content string) {
	if groupID == "" || messageID == "" || msgType == "" || content == "" {
		c.writeAck(groupAckMessage{Op: "ack", For: "send", Ok: false, Error: "missing_fields", MessageID: messageID})
		return
	}
	now := time.Now().UnixMilli()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if c.OnSend != nil {
		if err := c.OnSend(ctx, groupID, messageID, msgType, content, now); err != nil {
			c.writeAck(groupAckMessage{Op: "ack", For: "send", Ok: false, Error: err.Error(), GroupID: groupID, MessageID: messageID, ServerTS: now})
			return
		}
	}
	out, _ := json.Marshal(groupOutMessage{
		Op:      "new_message",
		GroupID: groupID,
		Message: groupMessageBody{
			MessageID:  messageID,
			FromUserID: c.UserID,
			Type:       msgType,
			Content:    content,
			Ctime:      now,
		},
	})
	c.Hub.Broadcast(groupID, out)
	c.writeAck(groupAckMessage{Op: "ack", For: "send", Ok: true, GroupID: groupID, MessageID: messageID, ServerTS: now})
}

func (c *GroupClient) writeAck(ack groupAckMessage) {
	b, err := json.Marshal(ack)
	if err != nil {
		return
	}
	select {
	case c.Send <- b:
	default:
	}
}

