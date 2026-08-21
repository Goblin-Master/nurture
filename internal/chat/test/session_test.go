package test

import (
	"nurture/internal/chat/constant"
	"nurture/internal/chat/session"
	"testing"
	"time"
)

func TestSessionHubSendToUserByChannel(t *testing.T) {
	hub := session.NewHub()
	go hub.Run()

	directClient := session.NewClient(hub, nil, "user-1", constant.ChannelDirect)
	groupClient := session.NewClient(hub, nil, "user-1", constant.ChannelGroup)
	hub.Register(directClient)
	hub.Register(groupClient)

	hub.SendToUser(constant.ChannelDirect, "user-1", []byte("direct-message"))

	got := readSessionMessage(t, directClient.Send)
	if string(got) != "direct-message" {
		t.Fatalf("got %q, want direct-message", got)
	}
	assertNoSessionMessage(t, groupClient.Send)
}

func TestSessionHubBroadcastToSubscribedRoom(t *testing.T) {
	hub := session.NewHub()
	go hub.Run()

	subscriber := session.NewClient(hub, nil, "user-1", constant.ChannelGroup)
	outsider := session.NewClient(hub, nil, "user-2", constant.ChannelGroup)
	hub.Register(subscriber)
	hub.Register(outsider)
	hub.Subscribe(subscriber, "group-1")

	hub.Broadcast("group-1", []byte("room-message"))

	got := readSessionMessage(t, subscriber.Send)
	if string(got) != "room-message" {
		t.Fatalf("got %q, want room-message", got)
	}
	assertNoSessionMessage(t, outsider.Send)
}

func TestSessionHubSkipsDuplicateDeliveryEvent(t *testing.T) {
	hub := session.NewHub()
	go hub.Run()

	client := session.NewClient(hub, nil, "user-1", constant.ChannelDirect)
	hub.Register(client)

	hub.DeliverToUser(constant.ChannelDirect, "user-1", "event-1", []byte("message-1"))
	hub.DeliverToUser(constant.ChannelDirect, "user-1", "event-1", []byte("message-1-duplicate"))

	got := readSessionMessage(t, client.Send)
	if string(got) != "message-1" {
		t.Fatalf("got %q, want message-1", got)
	}
	assertNoSessionMessage(t, client.Send)
}

func readSessionMessage(t *testing.T, ch <-chan []byte) []byte {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func assertNoSessionMessage(t *testing.T, ch <-chan []byte) {
	t.Helper()
	select {
	case got := <-ch:
		t.Fatalf("unexpected message %q", got)
	case <-time.After(50 * time.Millisecond):
	}
}
