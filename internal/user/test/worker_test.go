package test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"nurture/internal/pkg/rabbitmqx"
	user "nurture/internal/user"
	"nurture/internal/user/event"
	"nurture/internal/user/repo"
	"nurture/internal/user/worker"
)

type userOutboxConsumerFake struct {
	published []publishedUserEvent
	cancel    context.CancelFunc
}

type publishedUserEvent struct {
	exchange   string
	routingKey string
	messageID  string
	body       []byte
}

func (f *userOutboxConsumerFake) Publish(ctx context.Context, exchange, routingKey string, msg rabbitmqx.PublishMessage) error {
	f.published = append(f.published, publishedUserEvent{
		exchange:   exchange,
		routingKey: routingKey,
		messageID:  msg.MessageID,
		body:       append([]byte(nil), msg.Body...),
	})
	if f.cancel != nil {
		f.cancel()
	}
	return nil
}

func (f *userOutboxConsumerFake) DeclareTopicExchange(name string) error {
	return nil
}

type userOutboxRepoFake struct {
	items      []repo.UserOutboxEvent
	published  []int64
	failed     []int64
	nextCalled bool
}

func (f *userOutboxRepoFake) ListPendingOutbox(ctx context.Context, now int64, staleBefore int64, limit int) ([]repo.UserOutboxEvent, error) {
	return f.items, nil
}

func (f *userOutboxRepoFake) MarkOutboxPublished(ctx context.Context, id int64, now int64) error {
	f.published = append(f.published, id)
	return nil
}

func (f *userOutboxRepoFake) MarkOutboxFailed(ctx context.Context, id int64, nextRetryAt int64, maxAttempts int32, now int64) error {
	f.failed = append(f.failed, id)
	return nil
}

func TestUserWorkerPublishesPartnerBoundOutbox(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	msg, err := event.NewPartnerBoundMessage("father-1", "mother-1", 123)
	if err != nil {
		t.Fatalf("NewPartnerBoundMessage() error = %v", err)
	}
	repoFake := &userOutboxRepoFake{
		items: []repo.UserOutboxEvent{
			{
				ID:         7,
				EventID:    msg.EventID,
				RoutingKey: event.RoutingKeyPartnerBound,
				Payload:    msg.Payload,
				Attempts:   0,
				Ctime:      123,
			},
		},
	}
	bus := &userOutboxConsumerFake{cancel: cancel}
	w := worker.NewOutboxWorker(repoFake, bus, nil)

	w.Start(ctx)
	select {
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker publish")
	case <-ctx.Done():
	}

	if len(bus.published) != 1 {
		t.Fatalf("published len = %d, want 1", len(bus.published))
	}
	if bus.published[0].exchange != event.Exchange {
		t.Fatalf("publish exchange = %q, want %q", bus.published[0].exchange, event.Exchange)
	}
	if bus.published[0].routingKey != event.RoutingKeyPartnerBound {
		t.Fatalf("publish routing key = %q, want %q", bus.published[0].routingKey, event.RoutingKeyPartnerBound)
	}
	if bus.published[0].messageID != msg.EventID {
		t.Fatalf("publish message id = %q, want %q", bus.published[0].messageID, msg.EventID)
	}
	var decoded event.PartnerBoundMessage
	if err := json.Unmarshal(bus.published[0].body, &decoded); err != nil {
		t.Fatalf("published body json error: %v", err)
	}
	if decoded.EventID != msg.EventID || decoded.FatherUserID != msg.FatherUserID ||
		decoded.MotherUserID != msg.MotherUserID || decoded.OccurredAt != msg.OccurredAt {
		t.Fatalf("published body = %+v, want event fields from %+v", decoded, msg)
	}
	if len(repoFake.published) != 1 || repoFake.published[0] != 7 {
		t.Fatalf("published ids = %v, want [7]", repoFake.published)
	}
	if len(repoFake.failed) != 0 {
		t.Fatalf("failed ids = %v, want none", repoFake.failed)
	}
}

func TestUserModuleExposesOutboxWorkerPath(t *testing.T) {
	module := user.NewModule(user.Deps{
		Email: &emailFake{},
		SMS:   &smsFake{},
	})
	if module == nil {
		t.Fatal("module = nil")
	}
}

func TestUserOutboxWorkerRejectsNilDependencies(t *testing.T) {
	w := worker.NewOutboxWorker(nil, nil, nil)
	if w == nil {
		t.Fatal("NewOutboxWorker() = nil")
	}
	if err := w.Handle(context.Background(), repo.UserOutboxEvent{}); !errors.Is(err, worker.ErrNilOutboxEvent) {
		t.Fatalf("Handle() error = %v, want %v", err, worker.ErrNilOutboxEvent)
	}
}
