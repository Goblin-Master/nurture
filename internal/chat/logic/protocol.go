package logic

import (
	"context"
	"encoding/json"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/event"
	"time"
)

type GroupInMessage struct {
	Op        string   `json:"op"`
	GroupID   string   `json:"group_id"`
	GroupIDs  []string `json:"group_ids"`
	MessageID string   `json:"message_id"`
	Type      string   `json:"type"`
	Content   string   `json:"content"`
}

type GroupAckMessage struct {
	Op        string              `json:"op"`
	For       string              `json:"for"`
	Ok        bool                `json:"ok"`
	Error     string              `json:"error,omitempty"`
	GroupID   string              `json:"group_id,omitempty"`
	GroupIDs  []string            `json:"group_ids,omitempty"`
	Failed    []GroupFailedDetail `json:"failed,omitempty"`
	MessageID string              `json:"message_id,omitempty"`
	ServerTS  int64               `json:"server_ts,omitempty"`
}

type GroupFailedDetail struct {
	GroupID string `json:"group_id"`
	Error   string `json:"error"`
}

type GroupOutMessage struct {
	Op      string           `json:"op"`
	GroupID string           `json:"group_id"`
	Message GroupMessageBody `json:"message"`
}

type GroupMessageBody struct {
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}

type GroupMessageResult struct {
	Ack         *GroupAckMessage
	Subscribe   []string
	Unsubscribe []string
}

func (l *ChatLogic) HandleGroupMessage(ctx context.Context, userID string, message []byte) GroupMessageResult {
	var in GroupInMessage
	if err := json.Unmarshal(message, &in); err != nil {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "parse", Ok: false, Error: "invalid_json"}}
	}
	switch in.Op {
	case "subscribe":
		return l.handleGroupSubscribe(ctx, userID, in.GroupID)
	case "subscribe_many":
		return l.handleGroupSubscribeMany(ctx, userID, in.GroupIDs)
	case "unsubscribe":
		if in.GroupID == "" {
			return GroupMessageResult{}
		}
		return GroupMessageResult{Unsubscribe: []string{in.GroupID}}
	case "send":
		return l.handleGroupSend(ctx, userID, in.GroupID, in.MessageID, in.Type, in.Content)
	default:
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: in.Op, Ok: false, Error: "unknown_op"}}
	}
}

func (l *ChatLogic) handleGroupSubscribe(ctx context.Context, userID string, groupID string) GroupMessageResult {
	if groupID == "" {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: "missing_group_id"}}
	}
	if err := l.CheckMember(ctx, userID, groupID); err != nil {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "subscribe", Ok: false, Error: err.Error(), GroupID: groupID}}
	}
	return GroupMessageResult{
		Ack:       &GroupAckMessage{Op: "ack", For: "subscribe", Ok: true, GroupID: groupID},
		Subscribe: []string{groupID},
	}
}

func (l *ChatLogic) handleGroupSubscribeMany(ctx context.Context, userID string, groupIDs []string) GroupMessageResult {
	if len(groupIDs) == 0 {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "subscribe_many", Ok: false, Error: "missing_group_ids"}}
	}
	okIDs := make([]string, 0, len(groupIDs))
	failed := make([]GroupFailedDetail, 0)
	for _, groupID := range groupIDs {
		if groupID == "" {
			continue
		}
		if err := l.CheckMember(ctx, userID, groupID); err != nil {
			failed = append(failed, GroupFailedDetail{GroupID: groupID, Error: err.Error()})
			continue
		}
		okIDs = append(okIDs, groupID)
	}
	return GroupMessageResult{
		Ack:       &GroupAckMessage{Op: "ack", For: "subscribe_many", Ok: len(failed) == 0, GroupIDs: okIDs, Failed: failed},
		Subscribe: okIDs,
	}
}

func (l *ChatLogic) handleGroupSend(ctx context.Context, userID, groupID, messageID, msgType, content string) GroupMessageResult {
	if groupID == "" || messageID == "" || msgType == "" || content == "" {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "send", Ok: false, Error: "missing_fields", MessageID: messageID}}
	}
	now := time.Now().UnixMilli()
	if err := l.SaveMessage(ctx, userID, groupID, messageID, msgType, content, now); err != nil {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "send", Ok: false, Error: err.Error(), GroupID: groupID, MessageID: messageID, ServerTS: now}}
	}
	out, _ := json.Marshal(GroupOutMessage{
		Op:      "new_message",
		GroupID: groupID,
		Message: GroupMessageBody{
			MessageID:  messageID,
			FromUserID: userID,
			Type:       msgType,
			Content:    content,
			Ctime:      now,
		},
	})
	if err := l.publisher.PublishGroup(ctx, event.GroupMessage{
		EventID:    event.GroupEventID(groupID, messageID),
		MessageID:  messageID,
		GroupID:    groupID,
		FromUserID: userID,
		Type:       msgType,
		Content:    content,
		Ctime:      now,
		Payload:    string(out),
	}); err != nil {
		return GroupMessageResult{Ack: &GroupAckMessage{Op: "ack", For: "send", Ok: false, Error: ErrDefault.Error(), GroupID: groupID, MessageID: messageID, ServerTS: now}}
	}
	return GroupMessageResult{
		Ack: &GroupAckMessage{Op: "ack", For: "send", Ok: true, GroupID: groupID, MessageID: messageID, ServerTS: now},
	}
}

func IsGroupMessageType(msgType string) bool {
	switch msgType {
	case constant.MessageTypeText, constant.MessageTypeImage, constant.MessageTypeSystem:
		return true
	default:
		return false
	}
}
