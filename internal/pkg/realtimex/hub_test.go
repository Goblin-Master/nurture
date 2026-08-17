package realtimex

import (
	"testing"
	"time"
)

func TestHubSendToUserByChannel(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	directClient := NewClient(hub, nil, "user-1", ChannelDirect)
	groupClient := NewClient(hub, nil, "user-1", ChannelGroup)
	hub.Register(directClient)
	hub.Register(groupClient)

	hub.SendToUser(ChannelDirect, "user-1", []byte("direct-message"))

	got := readMessage(t, directClient.Send)
	if string(got) != "direct-message" {
		t.Fatalf("got %q, want direct-message", got)
	}
	assertNoMessage(t, groupClient.Send)
}

func TestHubBroadcastToSubscribedRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	subscriber := NewClient(hub, nil, "user-1", ChannelGroup)
	outsider := NewClient(hub, nil, "user-2", ChannelGroup)
	hub.Register(subscriber)
	hub.Register(outsider)
	hub.Subscribe(subscriber, "group-1")

	hub.Broadcast("group-1", []byte("room-message"))

	got := readMessage(t, subscriber.Send)
	if string(got) != "room-message" {
		t.Fatalf("got %q, want room-message", got)
	}
	assertNoMessage(t, outsider.Send)
}

func readMessage(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func assertNoMessage(t *testing.T, ch <-chan []byte) {
	t.Helper()
	select {
	case got := <-ch:
		t.Fatalf("unexpected message %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}
