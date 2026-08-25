package logic

import (
	"context"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/smsx"
	"nurture/internal/pkg/zapx"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"

	"go.uber.org/zap"
)

type BabySyncer interface {
	SyncPartnerBabies(ctx context.Context, fatherUserID string, motherUserID string) error
}

type IUserLogic interface {
	Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
	RegisterSMS(ctx context.Context, req dto.RegisterSMSReq) (dto.RegisterResp, error)
	GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetRegisterCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetRegisterSMSCode(ctx context.Context, req dto.GetSMSCodeReq) (dto.GetCodeResp, error)
	GetResetCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	ResetPassword(ctx context.Context, req dto.ResetPasswordReq) (dto.ResetPasswordResp, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateUserAdditionReq) (dto.UpdateUserAdditionResp, error)
	UpdateAvatar(ctx context.Context, userID string, req dto.UpdateAvatarReq) (dto.UpdateAvatarResp, error)
	GetBindPhoneCode(ctx context.Context, userID string, req dto.GetSMSCodeReq) (dto.GetCodeResp, error)
	BindPhone(ctx context.Context, userID string, req dto.BindPhoneReq) (dto.BindContactResp, error)
	GetBindEmailCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	BindEmail(ctx context.Context, userID string, req dto.BindEmailReq) (dto.BindContactResp, error)
	GetRebindPhoneCode(ctx context.Context, userID string, req dto.GetSMSCodeReq) (dto.GetCodeResp, error)
	RebindPhone(ctx context.Context, userID string, req dto.BindPhoneReq) (dto.BindContactResp, error)
	GetRebindEmailCode(ctx context.Context, userID string, req dto.GetCodeReq) (dto.GetCodeResp, error)
	RebindEmail(ctx context.Context, userID string, req dto.BindEmailReq) (dto.BindContactResp, error)
	BindPartner(ctx context.Context, userID string, req dto.PartnerBindReq) (dto.PartnerBindResp, error)
	GetPartner(ctx context.Context, userID string) (dto.PartnerGetResp, error)
	MyProfile(ctx context.Context, userID string) (dto.MyProfileResp, error)
	Follow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error)
	Unfollow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error)
	ListFollowing(ctx context.Context, userID string, req dto.FollowingListReq) (dto.FollowingListResp, error)
	ListFollowers(ctx context.Context, userID string, req dto.FollowersListReq) (dto.FollowersListResp, error)
	// admin
	AdminListUsers(ctx context.Context, req dto.AdminListUsersReq) (dto.AdminListUsersResp, error)
	AdminPromoteToAdmin(ctx context.Context, userID string) (string, error)
}

type UserLogic struct {
	userRepo   repo.IUserRepo
	email      emailx.Sender
	sms        smsx.Sender
	babySyncer BabySyncer
	log        *zap.SugaredLogger
}

func NewUserLogic(userRepo repo.IUserRepo, email emailx.Sender, sms smsx.Sender, babySyncer BabySyncer, log *zap.SugaredLogger) *UserLogic {
	return &UserLogic{
		userRepo:   userRepo,
		email:      email,
		sms:        sms,
		babySyncer: babySyncer,
		log:        zapx.OrNop(log),
	}
}

var _ IUserLogic = (*UserLogic)(nil)

func (ul *UserLogic) SetBabySyncer(syncer BabySyncer) {
	ul.babySyncer = syncer
}
