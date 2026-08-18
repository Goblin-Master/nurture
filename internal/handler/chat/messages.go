package chat

import (
	"context"
	"encoding/json"
	"time"
)

type groupInMessage struct {
	Op        string   `json:"op"`
	GroupID   string   `json:"group_id"`
	GroupIDs  []string `json:"group_ids"`
	MessageID string   `json:"message_id"`
	Type      string   `json:"type"`
	Content   string   `json:"content"`
}

type groupAckMessage struct {
	Op        string              `json:"op"`
	For       string              `json:"for"`
	Ok        bool                `json:"ok"`
	Error     string              `json:"error,omitempty"`
	GroupID   string              `json:"group_id,omitempty"`
	GroupIDs  []string            `json:"group_ids,omitempty"`
	Failed    []groupFailedDetail `json:"failed,omitempty"`
	MessageID string              `json:"message_id,omitempty"`
	ServerTS  int64               `json:"server_ts,omitempty"`
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

func (h *Handler) handleGroupMessage(client *session, message []byte) {
	var in groupInMessage
	if err := json.Unmarshal(message, &in); err != nil {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "parse", Ok: false, Error: "invalid_json"})
		return
	}
	switch in.Op {
	case "subscribe":
		h.handleGroupSubscribe(client, in.GroupID)
	case "subscribe_many":
		h.handleGroupSubscribeMany(client, in.GroupIDs)
	case "unsubscribe":
		if in.GroupID != "" {
			client.hub.unsubscribeGroup(client, in.GroupID)
		}
	case "send":
		h.handleGroupSend(client, in.GroupID, in.MessageID, in.Type, in.Content)
	default:
		writeGroupAck(client, groupAckMessage{Op: "ack", For: in.Op, Ok: false, Error: "unknown_op"})
	}
}

func (h *Handler) handleGroupSubscribe(client *session, groupID string) {
	if groupID == "" {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: "missing_group_id"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := h.chatLogic.CheckMember(ctx, client.userID, groupID); err != nil {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: err.Error(), GroupID: groupID})
		return
	}
	client.hub.subscribeGroup(client, groupID)
	writeGroupAck(client, groupAckMessage{Op: "ack", For: "subscribe", Ok: true, GroupID: groupID})
}

func (h *Handler) handleGroupSubscribeMany(client *session, groupIDs []string) {
	if len(groupIDs) == 0 {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "subscribe_many", Ok: false, Error: "missing_group_ids"})
		return
	}
	okIDs := make([]string, 0, len(groupIDs))
	failed := make([]groupFailedDetail, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, groupID := range groupIDs {
		if groupID == "" {
			continue
		}
		if err := h.chatLogic.CheckMember(ctx, client.userID, groupID); err != nil {
			failed = append(failed, groupFailedDetail{GroupID: groupID, Error: err.Error()})
			continue
		}
		client.hub.subscribeGroup(client, groupID)
		okIDs = append(okIDs, groupID)
	}
	writeGroupAck(client, groupAckMessage{Op: "ack", For: "subscribe_many", Ok: len(failed) == 0, GroupIDs: okIDs, Failed: failed})
}

func (h *Handler) handleGroupSend(client *session, groupID, messageID, msgType, content string) {
	if groupID == "" || messageID == "" || msgType == "" || content == "" {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "send", Ok: false, Error: "missing_fields", MessageID: messageID})
		return
	}
	now := time.Now().UnixMilli()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.chatLogic.SaveMessage(ctx, client.userID, groupID, messageID, msgType, content, now); err != nil {
		writeGroupAck(client, groupAckMessage{Op: "ack", For: "send", Ok: false, Error: err.Error(), GroupID: groupID, MessageID: messageID, ServerTS: now})
		return
	}
	out, _ := json.Marshal(groupOutMessage{
		Op:      "new_message",
		GroupID: groupID,
		Message: groupMessageBody{
			MessageID:  messageID,
			FromUserID: client.userID,
			Type:       msgType,
			Content:    content,
			Ctime:      now,
		},
	})
	client.hub.broadcastGroup(groupID, out)
	writeGroupAck(client, groupAckMessage{Op: "ack", For: "send", Ok: true, GroupID: groupID, MessageID: messageID, ServerTS: now})
}

func writeGroupAck(client *session, ack groupAckMessage) {
	b, err := json.Marshal(ack)
	if err != nil {
		return
	}
	client.trySend(b)
}
