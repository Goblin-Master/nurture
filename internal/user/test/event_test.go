package test

import (
	"encoding/json"
	"testing"

	"nurture/internal/user/event"
)

func TestPartnerBoundEventIDIncludesUsers(t *testing.T) {
	got := event.PartnerBoundEventID("father-id", "mother-id")
	want := "partner.bound:father-id:mother-id"
	if got != want {
		t.Fatalf("PartnerBoundEventID() = %q, want %q", got, want)
	}
}

func TestNewPartnerBoundMessageBuildsPayload(t *testing.T) {
	msg, err := event.NewPartnerBoundMessage("father-id", "mother-id", 123)
	if err != nil {
		t.Fatalf("NewPartnerBoundMessage() error = %v", err)
	}
	if msg.EventID != "partner.bound:father-id:mother-id" {
		t.Fatalf("EventID = %q, want partner.bound:father-id:mother-id", msg.EventID)
	}
	if msg.FatherUserID != "father-id" || msg.MotherUserID != "mother-id" || msg.OccurredAt != 123 {
		t.Fatalf("message = %+v, want father/mother/occurred_at", msg)
	}

	var decoded event.PartnerBoundMessage
	if err := json.Unmarshal([]byte(msg.Payload), &decoded); err != nil {
		t.Fatalf("payload json error: %v", err)
	}
	if decoded.EventID != msg.EventID || decoded.FatherUserID != msg.FatherUserID ||
		decoded.MotherUserID != msg.MotherUserID || decoded.OccurredAt != msg.OccurredAt {
		t.Fatalf("payload = %+v, want event fields from %+v", decoded, msg)
	}
}
