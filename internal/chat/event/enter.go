package event

import (
	"context"
	"encoding/json"
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

type Publisher interface {
	PublishDirect(ctx context.Context, msg DirectMessage) error
	PublishGroup(ctx context.Context, msg GroupMessage) error
}

type Bus interface {
	DeclareTopicExchange(name string) error
	Publish(ctx context.Context, exchange, routingKey string, msg rabbitmqx.PublishMessage) error
}

type Broker struct {
	bus Bus
}

type NoopPublisher struct{}

type DirectMessage struct {
	EventID    string `json:"event_id"`
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
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

func NewPublisher(bus Bus) (Publisher, error) {
	if bus == nil {
		return NoopPublisher{}, nil
	}
	if err := bus.DeclareTopicExchange(Exchange); err != nil {
		return nil, err
	}
	return &Broker{bus: bus}, nil
}

func (p NoopPublisher) PublishDirect(context.Context, DirectMessage) error {
	return nil
}

func (p NoopPublisher) PublishGroup(context.Context, GroupMessage) error {
	return nil
}

func (p *Broker) PublishDirect(ctx context.Context, msg DirectMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, Exchange, RoutingKeyDirect, rabbitmqx.PublishMessage{
		MessageID:   msg.EventID,
		ContentType: ContentTypeJSON,
		Body:        body,
	})
}

func (p *Broker) PublishGroup(ctx context.Context, msg GroupMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, Exchange, RoutingKeyGroup, rabbitmqx.PublishMessage{
		MessageID:   msg.EventID,
		ContentType: ContentTypeJSON,
		Body:        body,
	})
}

func DirectEventID(messageID string) string {
	return EventKindDirect + ":" + messageID
}

func GroupEventID(groupID string, messageID string) string {
	return EventKindGroup + ":" + groupID + ":" + messageID
}
