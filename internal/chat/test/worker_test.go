package test

import (
	"context"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/session"
	"nurture/internal/chat/worker"
	"nurture/internal/pkg/rabbitmqx"
	"testing"
	"time"
)

type chatConsumerFake struct {
	cfg    rabbitmqx.ConsumeConfig
	cancel context.CancelFunc
}

func (f *chatConsumerFake) Consume(_ context.Context, cfg rabbitmqx.ConsumeConfig, _ func(context.Context, rabbitmqx.Delivery) error) error {
	f.cfg = cfg
	if f.cancel != nil {
		f.cancel()
	}
	return context.Canceled
}

func TestWorkerEnablesConsumerRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	consumer := &chatConsumerFake{cancel: cancel}
	w := worker.NewWorker(consumer, session.NewHub(), nil)

	w.Start(ctx)
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker consume")
	}

	if consumer.cfg.Retry.Delay != constant.ConsumerRetryDelay {
		t.Fatalf("retry delay = %v, want %v", consumer.cfg.Retry.Delay, constant.ConsumerRetryDelay)
	}
	if consumer.cfg.Retry.MaxAttempts != constant.ConsumerMaxAttempts {
		t.Fatalf("retry max attempts = %d, want %d", consumer.cfg.Retry.MaxAttempts, constant.ConsumerMaxAttempts)
	}
}
