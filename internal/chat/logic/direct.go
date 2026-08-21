package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/dto"
	"nurture/internal/chat/event"
	"nurture/internal/chat/repo"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type DirectInMessage struct {
	Op        string `json:"op"`
	MessageID string `json:"message_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

type DirectAckMessage struct {
	Op        string `json:"op"`
	For       string `json:"for"`
	Ok        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	ServerTS  int64  `json:"server_ts,omitempty"`
}

type DirectOutMessage struct {
	Op      string            `json:"op"`
	Message DirectMessageBody `json:"message"`
}

type DirectMessageBody struct {
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}

type DirectMessageResult struct {
	Ack *DirectAckMessage
}

func (l *ChatLogic) HandleDirectMessage(ctx context.Context, userID string, partnerID string, message []byte) (DirectMessageResult, error) {
	userID = strings.TrimSpace(userID)
	partnerID = strings.TrimSpace(partnerID)

	var in DirectInMessage
	if err := json.Unmarshal(message, &in); err != nil {
		return directFailedAck("parse", "", "invalid_json", 0), nil
	}
	switch in.Op {
	case "send":
		return l.handleDirectSend(ctx, userID, partnerID, in), nil
	default:
		return directFailedAck(in.Op, in.MessageID, "unknown_op", 0), nil
	}
}

func (l *ChatLogic) handleDirectSend(ctx context.Context, userID string, partnerID string, in DirectInMessage) DirectMessageResult {
	content := normalizeDirectMessage([]byte(in.Content))
	now := time.Now().UnixMilli()
	if userID == "" || partnerID == "" || userID == partnerID || in.MessageID == "" || in.Type == "" || content == "" {
		return directFailedAck("send", in.MessageID, "missing_fields", now)
	}
	if _, err := uuid.Parse(in.MessageID); err != nil {
		return directFailedAck("send", in.MessageID, ErrParamsType.Error(), now)
	}
	if !IsChatMessageType(in.Type) {
		return directFailedAck("send", in.MessageID, ErrInvalidMessageType.Error(), now)
	}
	if l.limiter == nil {
		return directFailedAck("send", in.MessageID, ErrDefault.Error(), now)
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendUserKey, userID), constant.RateLimitWSSendUserLimit, constant.RateLimitWSSendWindow); !ok {
		return directFailedAck("send", in.MessageID, ErrTooManyRequests.Error(), now)
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendDirectKey, userID, partnerID), constant.RateLimitWSSendDirectLimit, constant.RateLimitWSSendWindow); !ok {
		return directFailedAck("send", in.MessageID, ErrTooManyRequests.Error(), now)
	}

	out, err := json.Marshal(DirectOutMessage{
		Op: "new_message",
		Message: DirectMessageBody{
			MessageID:  in.MessageID,
			FromUserID: userID,
			ToUserID:   partnerID,
			Type:       in.Type,
			Content:    content,
			Ctime:      now,
		},
	})
	if err != nil {
		return directFailedAck("send", in.MessageID, ErrDefault.Error(), now)
	}
	payload, err := json.Marshal(event.DirectMessage{
		EventID:    event.DirectEventID(in.MessageID),
		MessageID:  in.MessageID,
		FromUserID: userID,
		ToUserID:   partnerID,
		Type:       in.Type,
		Content:    content,
		Ctime:      now,
		Payload:    string(out),
	})
	if err != nil {
		return directFailedAck("send", in.MessageID, ErrDefault.Error(), now)
	}
	_, err = l.chatRepo.SaveDirectMessage(ctx, in.MessageID, userID, partnerID, in.Type, content, now, repo.ChatOutboxEvent{
		EventID:    event.DirectEventID(in.MessageID),
		RoutingKey: event.RoutingKeyDirect,
		Payload:    string(payload),
		Ctime:      now,
	})
	if err != nil {
		return directFailedAck("send", in.MessageID, mapRepoErr(err).Error(), now)
	}
	return DirectMessageResult{
		Ack: &DirectAckMessage{Op: "ack", For: "send", Ok: true, MessageID: in.MessageID, ServerTS: now},
	}
}

func (l *ChatLogic) ListDirectMessages(ctx context.Context, userID string, uri dto.ChatDirectMessageUserUri, req dto.ChatDirectMessageListReq) (dto.ChatDirectMessageListResp, error) {
	var resp dto.ChatDirectMessageListResp
	partnerID := strings.TrimSpace(uri.UserID)
	if strings.TrimSpace(userID) == "" || partnerID == "" || userID == partnerID {
		return resp, ErrParamsType
	}
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if req.Before != "" && req.After != "" {
		return resp, ErrParamsType
	}
	var rows []repo.ChatDirectMessageItem
	var err error
	if req.After != "" {
		ctime, mid, err := parseCursor(req.After)
		if err != nil {
			return resp, err
		}
		rows, err = l.chatRepo.ListDirectMessagesAfter(ctx, userID, partnerID, ctime, mid, limit)
		resp.NextAfter = nextDirectCursor(rows)
	} else if req.Before != "" {
		ctime, mid, err := parseCursor(req.Before)
		if err != nil {
			return resp, err
		}
		rows, err = l.chatRepo.ListDirectMessagesBefore(ctx, userID, partnerID, ctime, mid, limit)
		resp.NextBefore = nextDirectCursor(rows)
	} else {
		rows, err = l.chatRepo.ListDirectMessagesLatest(ctx, userID, partnerID, limit)
		resp.NextBefore = nextDirectCursor(rows)
	}
	if err != nil {
		return resp, mapRepoErr(err)
	}
	resp.Items = make([]dto.ChatDirectMessageItem, 0, len(rows))
	for _, v := range rows {
		resp.Items = append(resp.Items, dto.ChatDirectMessageItem{
			MessageID:  v.MessageID,
			FromUserID: v.FromUserID,
			ToUserID:   v.ToUserID,
			Type:       v.Type,
			Content:    v.Content,
			Ctime:      v.Ctime,
		})
	}
	resp.HasMore = len(resp.Items) >= limit
	return resp, nil
}

func (l *ChatLogic) MarkDirectSeen(ctx context.Context, userID string, partnerID string, now int64) error {
	partnerID = strings.TrimSpace(partnerID)
	if strings.TrimSpace(userID) == "" || partnerID == "" || userID == partnerID {
		return ErrParamsType
	}
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if err := l.chatRepo.MarkDirectSeen(ctx, userID, partnerID, now, now); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func directFailedAck(op string, messageID string, msg string, serverTS int64) DirectMessageResult {
	if op == "" {
		op = "send"
	}
	return DirectMessageResult{
		Ack: &DirectAckMessage{
			Op:        "ack",
			For:       op,
			Ok:        false,
			Error:     msg,
			MessageID: messageID,
			ServerTS:  serverTS,
		},
	}
}

func nextDirectCursor(rows []repo.ChatDirectMessageItem) string {
	if len(rows) == 0 {
		return ""
	}
	last := rows[len(rows)-1]
	return buildCursor(last.Ctime, last.MessageID)
}

func normalizeDirectMessage(message []byte) string {
	return string(bytes.TrimSpace(bytes.ReplaceAll(message, newline, space)))
}
