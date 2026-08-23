package event

import (
	"context"
	"nurture/internal/pkg/rabbitmqx"
)

const (
	Exchange         = "chat.event"
	RoutingKeyDirect = "chat.direct"
	RoutingKeyGroup  = "chat.group"
	ContentTypeJSON  = "application/json"
	ConsumerName     = "chat-worker"
	QueueNamePrefix  = "chat.event"
	EventKindDirect  = "direct"
	EventKindGroup   = "group"
)

type Bus interface {
	DeclareTopicExchange(name string) error
	Publish(ctx context.Context, exchange, routingKey string, msg rabbitmqx.PublishMessage) error
}

type DirectMessage struct {
	EventID    string `json:"event_id"`
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
	Payload    string `json:"payload"`
}

type GroupMessage struct {
	EventID    string `json:"event_id"`
	MessageID  string `json:"message_id"`
	GroupID    string `json:"group_id"`
	FromUserID string `json:"from_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
	Payload    string `json:"payload"`
}

func DirectEventID(fromUserID, toUserID, messageID string) string {
	return EventKindDirect + ":" + fromUserID + ":" + toUserID + ":" + messageID
}

func GroupEventID(groupID string, messageID string) string {
	return EventKindGroup + ":" + groupID + ":" + messageID
}
