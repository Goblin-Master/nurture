package test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"nurture/internal/baby/constant"
	babylogic "nurture/internal/baby/logic"
	"nurture/internal/baby/worker"
	"nurture/internal/pkg/rabbitmqx"
)

func TestBabyWorkerConsumesPartnerBoundEventWithRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	consumer := &babyConsumerFake{cancel: cancel}
	w := worker.NewWorker(consumer, &babyEventLogicFake{}, nil)

	w.Start(ctx)
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker consume")
	}

	if consumer.cfg.Exchange != constant.UserEventExchange {
		t.Fatalf("exchange = %q, want %q", consumer.cfg.Exchange, constant.UserEventExchange)
	}
	if consumer.cfg.Queue != constant.PartnerBoundQueue {
		t.Fatalf("queue = %q, want %q", consumer.cfg.Queue, constant.PartnerBoundQueue)
	}
	if !consumer.cfg.DurableQueue {
		t.Fatal("DurableQueue = false, want true")
	}
	if len(consumer.cfg.RoutingKeys) != 1 || consumer.cfg.RoutingKeys[0] != constant.UserPartnerBoundRoutingKey {
		t.Fatalf("routing keys = %v, want [%s]", consumer.cfg.RoutingKeys, constant.UserPartnerBoundRoutingKey)
	}
	if consumer.cfg.Retry.Delay != constant.ConsumerRetryDelay {
		t.Fatalf("retry delay = %v, want %v", consumer.cfg.Retry.Delay, constant.ConsumerRetryDelay)
	}
	if consumer.cfg.Retry.MaxAttempts != constant.ConsumerMaxAttempts {
		t.Fatalf("retry max attempts = %d, want %d", consumer.cfg.Retry.MaxAttempts, constant.ConsumerMaxAttempts)
	}
}

func TestBabyWorkerHandlesPartnerBoundEvent(t *testing.T) {
	eventID := "partner.bound:father-1:mother-1"
	body := mustJSON(t, map[string]any{
		"event_id":       eventID,
		"father_user_id": "father-1",
		"mother_user_id": "mother-1",
		"occurred_at":    int64(123),
	})
	logic := &babyEventLogicFake{}
	w := worker.NewWorker(nil, logic, nil)

	err := w.Handle(t.Context(), rabbitmqx.Delivery{
		MessageID:  eventID,
		RoutingKey: constant.UserPartnerBoundRoutingKey,
		Body:       body,
	})

	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if logic.eventID != eventID || logic.fatherID != "father-1" || logic.motherID != "mother-1" {
		t.Fatalf("Handle() got eventID=%q fatherID=%q motherID=%q", logic.eventID, logic.fatherID, logic.motherID)
	}
}

func TestBabyWorkerUsesDeliveryMessageIDWhenPayloadMissesEventID(t *testing.T) {
	body := mustJSON(t, map[string]any{
		"father_user_id": "father-1",
		"mother_user_id": "mother-1",
		"occurred_at":    int64(123),
	})
	logic := &babyEventLogicFake{}
	w := worker.NewWorker(nil, logic, nil)

	err := w.Handle(t.Context(), rabbitmqx.Delivery{
		MessageID:  "delivery-event-id",
		RoutingKey: constant.UserPartnerBoundRoutingKey,
		Body:       body,
	})

	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if logic.eventID != "delivery-event-id" {
		t.Fatalf("event id = %q, want delivery-event-id", logic.eventID)
	}
}

func TestBabyWorkerDiscardsBadPayload(t *testing.T) {
	w := worker.NewWorker(nil, &babyEventLogicFake{}, nil)

	err := w.Handle(t.Context(), rabbitmqx.Delivery{
		MessageID:  "bad-event",
		RoutingKey: constant.UserPartnerBoundRoutingKey,
		Body:       []byte("{"),
	})

	if err == nil || !strings.Contains(err.Error(), "decode partner bound event") {
		t.Fatalf("Handle() error = %v, want decode partner bound event error", err)
	}
}

func TestBabyWorkerDiscardsInvalidPartnerBoundEvent(t *testing.T) {
	w := worker.NewWorker(nil, &babyEventLogicFake{}, nil)

	err := w.Handle(t.Context(), rabbitmqx.Delivery{
		MessageID:  "bad-event",
		RoutingKey: constant.UserPartnerBoundRoutingKey,
		Body:       mustJSON(t, map[string]any{"father_user_id": "father-1"}),
	})

	if err == nil || !strings.Contains(err.Error(), "invalid partner bound event") {
		t.Fatalf("Handle() error = %v, want invalid partner bound event error", err)
	}
}

func TestBabyWorkerReturnsLogicErrorForRetry(t *testing.T) {
	wantErr := errors.New("repo unavailable")
	w := worker.NewWorker(nil, &babyEventLogicFake{err: wantErr}, nil)

	err := w.Handle(t.Context(), rabbitmqx.Delivery{
		MessageID:  "partner.bound:father-1:mother-1",
		RoutingKey: constant.UserPartnerBoundRoutingKey,
		Body: mustJSON(t, map[string]any{
			"father_user_id": "father-1",
			"mother_user_id": "mother-1",
		}),
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("Handle() error = %v, want %v", err, wantErr)
	}
}

func TestBabyLogicHandlesPartnerBoundInboxDuplicate(t *testing.T) {
	repo := &babyEventRepoFake{duplicate: true}
	l := babylogic.NewBabyEventLogic(repo, nil)

	err := l.HandlePartnerBound(t.Context(), "event-1", "father-1", "mother-1")

	if err != nil {
		t.Fatalf("HandlePartnerBound() error = %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repo calls = %d, want 1", repo.calls)
	}
}

func TestBabyLogicMapsPartnerSyncRepoError(t *testing.T) {
	repo := &babyEventRepoFake{err: errors.New("db down")}
	l := babylogic.NewBabyEventLogic(repo, nil)

	err := l.HandlePartnerBound(t.Context(), "event-1", "father-1", "mother-1")

	if !errors.Is(err, babylogic.ErrDefault) {
		t.Fatalf("HandlePartnerBound() error = %v, want %v", err, babylogic.ErrDefault)
	}
}

type babyConsumerFake struct {
	cfg    rabbitmqx.ConsumeConfig
	cancel context.CancelFunc
}

func (f *babyConsumerFake) Consume(_ context.Context, cfg rabbitmqx.ConsumeConfig, _ func(context.Context, rabbitmqx.Delivery) error) error {
	f.cfg = cfg
	if f.cancel != nil {
		f.cancel()
	}
	return context.Canceled
}

type babyEventLogicFake struct {
	eventID  string
	fatherID string
	motherID string
	err      error
}

func (f *babyEventLogicFake) HandlePartnerBound(ctx context.Context, eventID, fatherUserID, motherUserID string) error {
	f.eventID = eventID
	f.fatherID = fatherUserID
	f.motherID = motherUserID
	return f.err
}

type babyEventRepoFake struct {
	calls     int
	duplicate bool
	err       error
}

func (f *babyEventRepoFake) HandlePartnerBoundEvent(ctx context.Context, eventID, fatherUserID, motherUserID string) (bool, error) {
	f.calls++
	return !f.duplicate, f.err
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}
