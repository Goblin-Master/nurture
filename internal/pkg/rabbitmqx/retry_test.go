package rabbitmqx

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestConsumeTopologyWithRetry(t *testing.T) {
	cfg := ConsumeConfig{
		Exchange:    "chat.event",
		Queue:       "chat.event.node.1",
		RoutingKeys: []string{"chat.direct", "chat.group"},
		Retry: RetryConfig{
			Delay:       1500 * time.Millisecond,
			MaxAttempts: 3,
		},
	}

	topology := newConsumeTopology(cfg)
	if !topology.RetryEnabled {
		t.Fatal("retry topology disabled, want enabled")
	}
	if got, want := topology.RetryExchange, "chat.event.node.1.retry"; got != want {
		t.Fatalf("retry exchange = %q, want %q", got, want)
	}
	if got, want := topology.RetryQueue, "chat.event.node.1.retry"; got != want {
		t.Fatalf("retry queue = %q, want %q", got, want)
	}
	if got, want := topology.DeadExchange, "chat.event.node.1.dead"; got != want {
		t.Fatalf("dead exchange = %q, want %q", got, want)
	}
	if got, want := topology.DeadQueue, "chat.event.node.1.dead"; got != want {
		t.Fatalf("dead queue = %q, want %q", got, want)
	}
	if got, want := topology.MainQueueArgs["x-dead-letter-exchange"], topology.RetryExchange; got != want {
		t.Fatalf("main queue dlx = %v, want %v", got, want)
	}
	if got, want := topology.RetryQueueArgs["x-message-ttl"], int32(1500); got != want {
		t.Fatalf("retry queue ttl = %v, want %v", got, want)
	}
	if got, want := topology.RetryQueueArgs["x-dead-letter-exchange"], ""; got != want {
		t.Fatalf("retry queue dlx = %v, want %v", got, want)
	}
	if got, want := topology.RetryQueueArgs["x-dead-letter-routing-key"], cfg.Queue; got != want {
		t.Fatalf("retry queue dl routing key = %v, want %v", got, want)
	}
}

func TestConsumeTopologyRetryDisabledByDefault(t *testing.T) {
	topology := newConsumeTopology(ConsumeConfig{Queue: "chat.event.node.1"})
	if topology.RetryEnabled {
		t.Fatal("retry topology enabled, want disabled")
	}
	if topology.MainQueueArgs != nil {
		t.Fatalf("main queue args = %#v, want nil", topology.MainQueueArgs)
	}
}

func TestDeliveryAttemptUsesXDeathForMainQueue(t *testing.T) {
	headers := amqp.Table{
		"x-death": []any{
			amqp.Table{"queue": "chat.event.node.1.retry", "count": int64(2)},
			amqp.Table{"queue": "chat.event.node.1", "count": int64(2)},
		},
	}

	got := deliveryAttempt(headers, "chat.event.node.1")
	if got != 3 {
		t.Fatalf("deliveryAttempt() = %d, want 3", got)
	}
}
