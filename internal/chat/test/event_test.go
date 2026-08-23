package test

import (
	"nurture/internal/chat/event"
	"testing"
)

func TestDirectEventIDIncludesSenderAndReceiver(t *testing.T) {
	got := event.DirectEventID("sender-id", "receiver-id", "message-id")
	want := "direct:sender-id:receiver-id:message-id"
	if got != want {
		t.Fatalf("DirectEventID() = %q, want %q", got, want)
	}
}

func TestGroupEventIDIncludesGroupAndMessage(t *testing.T) {
	got := event.GroupEventID("group-id", "message-id")
	want := "group:group-id:message-id"
	if got != want {
		t.Fatalf("GroupEventID() = %q, want %q", got, want)
	}
}
