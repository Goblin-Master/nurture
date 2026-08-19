package logic

import (
	"bytes"
	"context"
	"fmt"
	"nurture/internal/chat/constant"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type DirectMessageResult struct {
	RecipientID string
	Message     []byte
}

func (l *ChatLogic) HandleDirectMessage(ctx context.Context, userID string, partnerID string, message []byte) (DirectMessageResult, error) {
	userID = strings.TrimSpace(userID)
	partnerID = strings.TrimSpace(partnerID)
	content := normalizeDirectMessage(message)
	if userID == "" || partnerID == "" || userID == partnerID || content == "" {
		return DirectMessageResult{}, ErrParamsType
	}
	if l.limiter == nil {
		return DirectMessageResult{}, ErrDefault
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendUserKey, userID), constant.RateLimitWSSendUserLimit, constant.RateLimitWSSendWindow); !ok {
		return DirectMessageResult{}, ErrTooManyRequests
	}
	if ok, _, _ := l.limiter.Allow(ctx, fmt.Sprintf(constant.RateLimitWSSendDirectKey, userID, partnerID), constant.RateLimitWSSendDirectLimit, constant.RateLimitWSSendWindow); !ok {
		return DirectMessageResult{}, ErrTooManyRequests
	}

	messageID := uuid.NewString()
	now := time.Now().UnixMilli()
	if err := l.chatRepo.SaveDirectMessage(ctx, messageID, userID, partnerID, constant.MessageTypeText, content, now); err != nil {
		return DirectMessageResult{}, mapRepoErr(err)
	}
	return DirectMessageResult{
		RecipientID: partnerID,
		Message:     []byte(content),
	}, nil
}

func normalizeDirectMessage(message []byte) string {
	return string(bytes.TrimSpace(bytes.ReplaceAll(message, newline, space)))
}
