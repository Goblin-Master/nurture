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
	if topology.RetryQueueOptions.durable {
		t.Fatal("retry queue durable = true, want false for ephemeral consumer")
	}
	if !topology.RetryQueueOptions.autoDelete {
		t.Fatal("retry queue autoDelete = false, want true for ephemeral consumer")
	}
	if !topology.RetryQueueOptions.exclusive {
		t.Fatal("retry queue exclusive = false, want true for ephemeral consumer")
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

func TestConsumeTopologyUsesDurableRetryQueuesForDurableConsumer(t *testing.T) {
	topology := newConsumeTopology(ConsumeConfig{
		Queue:        "baby.partner.bound",
		DurableQueue: true,
		Retry: RetryConfig{
			Delay:       time.Second,
			MaxAttempts: 3,
		},
	})

	if !topology.RetryQueueOptions.durable {
		t.Fatal("retry queue durable = false, want true")
	}
	if topology.RetryQueueOptions.autoDelete {
		t.Fatal("retry queue autoDelete = true, want false")
	}
	if topology.RetryQueueOptions.exclusive {
		t.Fatal("retry queue exclusive = true, want false")
	}
	if !topology.DeadQueueOptions.durable {
		t.Fatal("dead queue durable = false, want true")
	}
	if topology.DeadQueueOptions.autoDelete {
		t.Fatal("dead queue autoDelete = true, want false")
	}
	if topology.DeadQueueOptions.exclusive {
		t.Fatal("dead queue exclusive = true, want false")
	}
}

func TestConsumeQueueOptionsDefaultToEphemeralExclusive(t *testing.T) {
	options := consumeQueueOptions(ConsumeConfig{Queue: "chat.event.node.1"})

	if options.durable {
		t.Fatal("durable = true, want false")
	}
	if !options.autoDelete {
		t.Fatal("autoDelete = false, want true")
	}
	if !options.exclusive {
		t.Fatal("exclusive = false, want true")
	}
}

func TestConsumeQueueOptionsSupportDurableSharedQueue(t *testing.T) {
	options := consumeQueueOptions(ConsumeConfig{Queue: "baby.partner.bound", DurableQueue: true})

	if !options.durable {
		t.Fatal("durable = false, want true")
	}
	if options.autoDelete {
		t.Fatal("autoDelete = true, want false")
	}
	if options.exclusive {
		t.Fatal("exclusive = true, want false")
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
