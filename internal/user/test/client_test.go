package test

import (
	"context"
	"testing"

	user "nurture/internal/user"
	userdto "nurture/internal/user/dto"
	userlogic "nurture/internal/user/logic"
	userrepo "nurture/internal/user/repo"
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

func TestUserModuleExposesClientAndAcceptsLateBabySyncer(t *testing.T) {
	module := user.NewModule(user.Deps{
		Email: &emailFake{},
		SMS:   &smsFake{},
	})
	if module.Client() == nil {
		t.Fatal("Client() = nil, want non-nil")
	}

	module.SetBabySyncer(&babySyncerFake{})
}

func TestUserLogicSetBabySyncerUpdatesPartnerBindingBoundary(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	partnerID := "22222222-2222-2222-2222-222222222222"
	repo := &userRepoFake{
		loginWithAccountRow: userrepo.UserBaseRow{UserID: partnerID, Gender: "female"},
		getUserByIDRow:      userrepo.UserBaseRow{UserID: userID, Gender: "male"},
		profileRows: map[string]userrepo.ProfileRow{
			partnerID: {UserID: partnerID, Username: "Partner", Avatar: "avatar.png"},
		},
	}
	logic := userlogic.NewUserLogic(repo, &emailFake{}, &smsFake{}, nil, nil)
	syncer := &babySyncerFake{}
	logic.SetBabySyncer(syncer)

	_, err := logic.BindPartner(context.Background(), userID, userdto.PartnerBindReq{
		Account:  "partner",
		Password: "Aa123456",
	})
	if err != nil {
		t.Fatalf("BindPartner() error = %v", err)
	}
	if syncer.fatherID != userID || syncer.motherID != partnerID {
		t.Fatalf("SyncPartnerBabies() got (%q,%q), want (%q,%q)", syncer.fatherID, syncer.motherID, userID, partnerID)
	}
}
