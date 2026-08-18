package test

import (
	"nurture/internal/chat/constant"
	"nurture/internal/chat/session"
	"testing"
	"time"
)

func TestRealtimeHubSendToUserByChannel(t *testing.T) {
	hub := session.NewHub()
	go hub.Run()

	directClient := session.NewClient(hub, nil, "user-1", constant.ChannelDirect)
	groupClient := session.NewClient(hub, nil, "user-1", constant.ChannelGroup)
	hub.Register(directClient)
	hub.Register(groupClient)

	hub.SendToUser(constant.ChannelDirect, "user-1", []byte("direct-message"))

	got := readRealtimeMessage(t, directClient.Send)
	if string(got) != "direct-message" {
		t.Fatalf("got %q, want direct-message", got)
	}
	assertNoRealtimeMessage(t, groupClient.Send)
}

func TestRealtimeHubBroadcastToSubscribedRoom(t *testing.T) {
	hub := session.NewHub()
	go hub.Run()

	subscriber := session.NewClient(hub, nil, "user-1", constant.ChannelGroup)
	outsider := session.NewClient(hub, nil, "user-2", constant.ChannelGroup)
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
