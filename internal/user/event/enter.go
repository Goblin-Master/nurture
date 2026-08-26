package event

import (
	"encoding/json"
	"fmt"
)

const (
	Exchange               = "user.event"
	RoutingKeyPartnerBound = "partner.bound"
	ContentTypeJSON        = "application/json"
)

type PartnerBoundMessage struct {
	EventID      string `json:"event_id"`
	FatherUserID string `json:"father_user_id"`
	MotherUserID string `json:"mother_user_id"`
	OccurredAt   int64  `json:"occurred_at"`
	Payload      string `json:"-"`
}

func PartnerBoundEventID(fatherUserID, motherUserID string) string {
	return fmt.Sprintf("%s:%s:%s", RoutingKeyPartnerBound, fatherUserID, motherUserID)
}

func NewPartnerBoundMessage(fatherUserID, motherUserID string, occurredAt int64) (PartnerBoundMessage, error) {
	msg := PartnerBoundMessage{
		EventID:      PartnerBoundEventID(fatherUserID, motherUserID),
		FatherUserID: fatherUserID,
		MotherUserID: motherUserID,
		OccurredAt:   occurredAt,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return PartnerBoundMessage{}, err
	}
	msg.Payload = string(payload)
	return msg, nil
}
