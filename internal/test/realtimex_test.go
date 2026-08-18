package test

import (
	"nurture/internal/pkg/realtimex"
	"testing"
	"time"
)

func TestRealtimeHubSendToUserByChannel(t *testing.T) {
	hub := realtimex.NewHub()
	go hub.Run()

	directClient := realtimex.NewClient(hub, nil, "user-1", realtimex.ChannelDirect)
	groupClient := realtimex.NewClient(hub, nil, "user-1", realtimex.ChannelGroup)
	hub.Register(directClient)
	hub.Register(groupClient)

	hub.SendToUser(realtimex.ChannelDirect, "user-1", []byte("direct-message"))

	got := readRealtimeMessage(t, directClient.Send)
	if string(got) != "direct-message" {
		t.Fatalf("got %q, want direct-message", got)
	}
	assertNoRealtimeMessage(t, groupClient.Send)
}

func TestRealtimeHubBroadcastToSubscribedRoom(t *testing.T) {
	hub := realtimex.NewHub()
	go hub.Run()

	subscriber := realtimex.NewClient(hub, nil, "user-1", realtimex.ChannelGroup)
	outsider := realtimex.NewClient(hub, nil, "user-2", realtimex.ChannelGroup)
	hub.Register(subscriber)
	hub.Register(outsider)
	hub.Subscribe(subscriber, "group-1")

	hub.Broadcast("group-1", []byte("room-message"))

	got := readRealtimeMessage(t, subscriber.Send)
	if string(got) != "room-message" {
		t.Fatalf("got %q, want room-message", got)
	}
	assertNoRealtimeMessage(t, outsider.Send)
}

func readRealtimeMessage(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func assertNoRealtimeMessage(t *testing.T, ch <-chan []byte) {
	t.Helper()
	select {
	case got := <-ch:
		t.Fatalf("unexpected message %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}
