package test

import (
	"context"
	"testing"

	user "nurture/internal/user"
)

func TestUserClientDelegatesPartnerAndFollowReads(t *testing.T) {
	repo := &userRepoFake{
		partnerID:   "partner-1",
		isFollowing: true,
	}
	client := user.NewClient(repo)

	partnerID, err := client.GetPartnerByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetPartnerByUserID() error = %v", err)
	}
	if partnerID != "partner-1" {
		t.Fatalf("GetPartnerByUserID() = %q, want partner-1", partnerID)
	}

	ok, err := client.IsFollowing(context.Background(), "user-1", "user-2")
	if err != nil {
		t.Fatalf("IsFollowing() error = %v", err)
	}
	if !ok {
		t.Fatal("IsFollowing() = false, want true")
	}
}

func TestUserModuleExposesClient(t *testing.T) {
	module := user.NewModule(user.Deps{
		Email: &emailFake{},
		SMS:   &smsFake{},
	})
	if module.Client() == nil {
		t.Fatal("Client() = nil, want non-nil")
	}
}
