package test

import (
	"context"
	"errors"
	"fmt"
	userconstant "nurture/internal/user/constant"
	userdto "nurture/internal/user/dto"
	userlogic "nurture/internal/user/logic"
	userrepo "nurture/internal/user/repo"
	"strings"
	"testing"
)

func TestRegisterMapsEmailConflict(t *testing.T) {
	email := &emailFake{verifyOK: true}
	repo := &userRepoFake{registerErr: userrepo.ErrEmailIsUsed}
	l := userlogic.NewUserLogic(repo, email, &smsFake{}, nil, nil)

	req := userdto.RegisterReq{
		Account:  "account",
		Password: "Aa123456",
		Username: "Alice",
		Gender:   "female",
		Email:    "alice@example.com",
		Code:     "123456",
	}
	_, err := l.Register(context.Background(), req)

	if !errors.Is(err, userlogic.ErrEmailIsUsed) {
		t.Fatalf("Register() error = %v, want %v", err, userlogic.ErrEmailIsUsed)
	}
	wantKey := fmt.Sprintf(userconstant.RegisterCodeKey, req.Email)
	if email.verifyKey != wantKey {
		t.Fatalf("VerifyCode() key = %q, want %q", email.verifyKey, wantKey)
	}
}

func TestGetRegisterCodeUsesUserCodeKey(t *testing.T) {
	email := &emailFake{}
	l := userlogic.NewUserLogic(&userRepoFake{}, email, &smsFake{}, nil, nil)

	req := userdto.GetCodeReq{Email: "alice@example.com"}
	resp, err := l.GetRegisterCode(context.Background(), req)

	if err != nil {
		t.Fatalf("GetRegisterCode() error = %v", err)
	}
	if len(resp.Code) != 6 {
		t.Fatalf("GetRegisterCode() code length = %d, want 6", len(resp.Code))
	}
	wantKey := fmt.Sprintf(userconstant.RegisterCodeKey, req.Email)
	if email.sendKey != wantKey {
		t.Fatalf("SendCode() key = %q, want %q", email.sendKey, wantKey)
	}
	if email.sendTitle != "注册账号" {
		t.Fatalf("SendCode() title = %q, want 注册账号", email.sendTitle)
	}
	if email.sendCode != resp.Code || !strings.Contains(email.sendText, resp.Code) {
		t.Fatalf("SendCode() did not receive generated code")
	}
}

func TestBindPartnerSyncsBabiesThroughInjectedBoundary(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	partnerID := "22222222-2222-2222-2222-222222222222"
	repo := &userRepoFake{
		loginWithAccountRow: userrepo.UserBaseRow{UserID: partnerID, Gender: "female"},
		getUserByIDRow:      userrepo.UserBaseRow{UserID: userID, Gender: "male"},
		profileRows: map[string]userrepo.ProfileRow{
			partnerID: {UserID: partnerID, Username: "Partner", Avatar: "avatar.png"},
		},
	}
	syncer := &babySyncerFake{}
	l := userlogic.NewUserLogic(repo, &emailFake{}, &smsFake{}, syncer, nil)

	resp, err := l.BindPartner(context.Background(), userID, userdto.PartnerBindReq{
		Account:  "partner",
		Password: "Aa123456",
	})

	if err != nil {
		t.Fatalf("BindPartner() error = %v", err)
	}
	if repo.boundFatherID != userID || repo.boundMotherID != partnerID {
		t.Fatalf("BindPartner() bound (%q,%q), want (%q,%q)", repo.boundFatherID, repo.boundMotherID, userID, partnerID)
	}
	if syncer.fatherID != userID || syncer.motherID != partnerID {
		t.Fatalf("SyncPartnerBabies() got (%q,%q), want (%q,%q)", syncer.fatherID, syncer.motherID, userID, partnerID)
	}
	if resp.PartnerID != partnerID || resp.PartnerUsername != "Partner" {
		t.Fatalf("BindPartner() resp = %+v", resp)
	}
}

func TestBindPartnerRejectsDifferentExistingPartner(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	partnerID := "22222222-2222-2222-2222-222222222222"
	repo := &userRepoFake{
		loginWithAccountRow: userrepo.UserBaseRow{UserID: partnerID, Gender: "female"},
		getUserByIDRow:      userrepo.UserBaseRow{UserID: userID, Gender: "male"},
		partnerID:           "33333333-3333-3333-3333-333333333333",
	}
	syncer := &babySyncerFake{}
	l := userlogic.NewUserLogic(repo, &emailFake{}, &smsFake{}, syncer, nil)

	_, err := l.BindPartner(context.Background(), userID, userdto.PartnerBindReq{
		Account:  "partner",
		Password: "Aa123456",
	})

	if !errors.Is(err, userlogic.ErrPartnerAlreadyBound) {
		t.Fatalf("BindPartner() error = %v, want %v", err, userlogic.ErrPartnerAlreadyBound)
	}
	if repo.boundFatherID != "" || syncer.fatherID != "" {
		t.Fatalf("BindPartner() should not bind or sync when an existing different partner is present")
	}
}

func TestUpdateProfileMapsUpdateFailure(t *testing.T) {
	repo := &userRepoFake{updateAdditionErr: userrepo.ErrUserUpdateFailed}
	l := userlogic.NewUserLogic(repo, &emailFake{}, &smsFake{}, nil, nil)
	occupation := "engineer"

	_, err := l.UpdateProfile(context.Background(), "user-id", userdto.UpdateUserAdditionReq{
		Occupation: &occupation,
	})

	if !errors.Is(err, userlogic.ErrProfileUpdateFailed) {
		t.Fatalf("UpdateProfile() error = %v, want %v", err, userlogic.ErrProfileUpdateFailed)
	}
}

type emailFake struct {
	verifyOK  bool
	verifyErr error
	verifyKey string
	sendErr   error
	sendTo    string
	sendTitle string
	sendText  string
	sendKey   string
	sendCode  string
}

func (f *emailFake) SendCode(ctx context.Context, to string, title string, text string, key string, code string) error {
	f.sendTo = to
	f.sendTitle = title
	f.sendText = text
	f.sendKey = key
	f.sendCode = code
	return f.sendErr
}

func (f *emailFake) VerifyCode(ctx context.Context, key string, code string) (bool, error) {
	f.verifyKey = key
	return f.verifyOK, f.verifyErr
}

type smsFake struct {
	verifyOK  bool
	verifyErr error
	sendErr   error
}

func (f *smsFake) SendCode(ctx context.Context, key string, phone string, code string) error {
	return f.sendErr
}

func (f *smsFake) VerifyCode(ctx context.Context, key string, code string) (bool, error) {
	return f.verifyOK, f.verifyErr
}

type babySyncerFake struct {
	fatherID string
	motherID string
	err      error
}

func (f *babySyncerFake) SyncPartnerBabies(ctx context.Context, fatherUserID string, motherUserID string) error {
	f.fatherID = fatherUserID
	f.motherID = motherUserID
	return f.err
}

type userRepoFake struct {
	loginWithAccountRow userrepo.UserBaseRow
	loginWithAccountErr error
	loginWithEmailRow   userrepo.UserBaseRow
	loginWithEmailErr   error
	getUserByIDRow      userrepo.UserBaseRow
	getUserByIDErr      error
	registerErr         error
	resetPasswordErr    error
	updateAvatarErr     error
	updateAdditionErr   error
	updateGenderErr     error
	profileRow          userrepo.ProfileRow
	profileRows         map[string]userrepo.ProfileRow
	profileErr          error
	partnerID           string
	partnerErr          error
	bindPartnerErr      error
	boundFatherID       string
	boundMotherID       string
	followErr           error
	unfollowErr         error
	isFollowing         bool
	isFollowingErr      error
	isPhoneUsed         bool
	isPhoneUsedErr      error
	bindEmailErr        error
	adminListRows       []userrepo.AdminUserRow
	adminListHasMore    bool
	adminListErr        error
	adminUpdateRoleErr  error
}

func (f *userRepoFake) LoginWithAccount(ctx context.Context, account string, password string) (userrepo.UserBaseRow, error) {
	return f.loginWithAccountRow, f.loginWithAccountErr
}

func (f *userRepoFake) LoginWithEmail(ctx context.Context, email string) (userrepo.UserBaseRow, error) {
	return f.loginWithEmailRow, f.loginWithEmailErr
}

func (f *userRepoFake) GetUserByID(ctx context.Context, userID string) (userrepo.UserBaseRow, error) {
	return f.getUserByIDRow, f.getUserByIDErr
}

func (f *userRepoFake) Register(ctx context.Context, userID, username string, email *string, account, password, gender string) error {
	return f.registerErr
}

func (f *userRepoFake) ResetPassword(ctx context.Context, email, newPassword string) error {
	return f.resetPasswordErr
}

func (f *userRepoFake) UpdateAvatarByID(ctx context.Context, userID, url string) error {
	return f.updateAvatarErr
}

func (f *userRepoFake) UpdateAdditionByID(ctx context.Context, userID string, occupation, phone, province, city, avatar *string, birthday *int64) error {
	return f.updateAdditionErr
}

func (f *userRepoFake) GetMyProfile(ctx context.Context, userID string) (userrepo.ProfileRow, error) {
	if f.profileRows != nil {
		if row, ok := f.profileRows[userID]; ok {
			return row, f.profileErr
		}
	}
	return f.profileRow, f.profileErr
}

func (f *userRepoFake) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	return f.partnerID, f.partnerErr
}

func (f *userRepoFake) BindPartner(ctx context.Context, fatherUserID, motherUserID string) error {
	f.boundFatherID = fatherUserID
	f.boundMotherID = motherUserID
	return f.bindPartnerErr
}

func (f *userRepoFake) UpdateGender(ctx context.Context, userID, gender string) error {
	return f.updateGenderErr
}

func (f *userRepoFake) FollowUser(ctx context.Context, followerID, followeeID string) error {
	return f.followErr
}

func (f *userRepoFake) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	return f.unfollowErr
}

func (f *userRepoFake) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]userrepo.FollowUserRow, bool, error) {
	return nil, false, nil
}

func (f *userRepoFake) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]userrepo.FollowUserRow, bool, error) {
	return nil, false, nil
}

func (f *userRepoFake) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return f.isFollowing, f.isFollowingErr
}

func (f *userRepoFake) IsPhoneUsed(ctx context.Context, phone string, excludeUserID string) (bool, error) {
	return f.isPhoneUsed, f.isPhoneUsedErr
}

func (f *userRepoFake) BindEmail(ctx context.Context, userID, email string) error {
	return f.bindEmailErr
}

func (f *userRepoFake) AdminListUsers(ctx context.Context, keyword string, page, pageSize int) ([]userrepo.AdminUserRow, bool, error) {
	return f.adminListRows, f.adminListHasMore, f.adminListErr
}

func (f *userRepoFake) AdminUpdateUserRole(ctx context.Context, userID string, role int16) error {
	return f.adminUpdateRoleErr
}
