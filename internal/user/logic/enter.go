package logic

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/passwordx"
	"nurture/internal/pkg/smsx"
	"nurture/internal/pkg/zapx"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (ul *UserLogic) Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error) {
	var resp dto.LoginResp
	switch req.LoginType {
	case userconstant.LoginWithAccount:
		data, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
		if err != nil {
			return resp, ErrAccountOrPassword
		}
		token, err := jwtx.GenToken(jwtx.Claims{
			UserID: data.UserID,
			Role:   jwtx.Role(data.Role),
		})
		if err != nil {
			ul.log.Error(err)
			return resp, ErrDefault
		}
		resp.Token = token
		return resp, nil
	case userconstant.LoginWithEmail:
		ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(userconstant.LoginCodeKey, req.Email), req.Code)
		if err != nil {
			ul.log.Error(err)
			return resp, ErrCodeVerify
		}
		if !ok {
			return resp, ErrCodeVerify
		}
		data, err := ul.userRepo.LoginWithEmail(ctx, req.Email)
		if err != nil {
			return resp, ErrEmail
		}
		token, err := jwtx.GenToken(jwtx.Claims{
			UserID: data.UserID,
			Role:   jwtx.Role(data.Role),
		})
		if err != nil {
			ul.log.Error(err)
			return resp, ErrDefault
		}
		resp.Token = token
		return resp, nil
	default:
		ul.log.Warnf("错误的登录方式:%s", req.LoginType)
		return resp, ErrLoginWithFailedWay
	}
}

func (ul *UserLogic) Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error) {
	var resp dto.RegisterResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(userconstant.RegisterCodeKey, req.Email), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	email := req.Email
	err = ul.userRepo.Register(ctx, uuid.NewString(), req.Username, &email, req.Account, req.Password, req.Gender)
	if err != nil {
		if errors.Is(err, passwordx.ErrPasswordEmpty) {
			return resp, ErrPasswordEmpty
		} else if errors.Is(err, passwordx.ErrPasswordTooShort) {
			return resp, ErrPasswordTooShort
		} else if errors.Is(err, passwordx.ErrPasswordTooLong) {
			return resp, ErrPasswordTooLong
		} else if errors.Is(err, passwordx.ErrPasswordTooWeak) {
			return resp, ErrPasswordTooWeak
		}
		if errors.Is(err, repo.ErrEmailIsUsed) {
			return resp, ErrEmailIsUsed
		} else if errors.Is(err, repo.ErrAccountIsUsed) {
			return resp, ErrAccountIsUsed
		} else {
			ul.log.Error(err)
			return resp, ErrDefault
		}
	}
	resp.Message = "用户注册成功！"
	return resp, nil
}

func (ul *UserLogic) RegisterSMS(ctx context.Context, req dto.RegisterSMSReq) (dto.RegisterResp, error) {
	var resp dto.RegisterResp
	if req.Gender != "male" && req.Gender != "female" {
		return resp, ErrInvalidGender
	}
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = phone
	}
	ok, err := ul.sms.VerifyCode(ctx, fmt.Sprintf(userconstant.RegisterSMSCodeKey, phone), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	userID := uuid.NewString()
	err = ul.userRepo.Register(ctx, userID, req.Username, nil, account, req.Password, req.Gender)
	if err != nil {
		if errors.Is(err, passwordx.ErrPasswordEmpty) {
			return resp, ErrPasswordEmpty
		} else if errors.Is(err, passwordx.ErrPasswordTooShort) {
			return resp, ErrPasswordTooShort
		} else if errors.Is(err, passwordx.ErrPasswordTooLong) {
			return resp, ErrPasswordTooLong
		} else if errors.Is(err, passwordx.ErrPasswordTooWeak) {
			return resp, ErrPasswordTooWeak
		}
		if errors.Is(err, repo.ErrAccountIsUsed) {
			return resp, ErrAccountIsUsed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	_ = ul.userRepo.UpdateAdditionByID(ctx, userID, nil, &phone, nil, nil, nil, nil)
	resp.Message = "用户注册成功！"
	return resp, nil
}

func (ul *UserLogic) ResetPassword(ctx context.Context, req dto.ResetPasswordReq) (dto.ResetPasswordResp, error) {
	var resp dto.ResetPasswordResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(userconstant.ResetPwdCodeKey, req.Email), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	err = ul.userRepo.ResetPassword(ctx, req.Email, req.NewPassword)
	if err != nil {
		if errors.Is(err, passwordx.ErrPasswordEmpty) {
			return resp, ErrPasswordEmpty
		} else if errors.Is(err, passwordx.ErrPasswordTooShort) {
			return resp, ErrPasswordTooShort
		} else if errors.Is(err, passwordx.ErrPasswordTooLong) {
			return resp, ErrPasswordTooLong
		} else if errors.Is(err, passwordx.ErrPasswordTooWeak) {
			return resp, ErrPasswordTooWeak
		}
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "重置密码成功！"
	return resp, nil
}

func (ul *UserLogic) GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c, err := emailx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.email.SendCode(
		ctx,
		req.Email,
		"邮箱登录",
		fmt.Sprintf("你正在进行邮箱登录，登录的验证码是：%s，十分钟内有效", c),
		fmt.Sprintf(userconstant.LoginCodeKey, req.Email),
		c,
	)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetRegisterCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c, err := emailx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.email.SendCode(
		ctx,
		req.Email,
		"注册账号",
		fmt.Sprintf("你正在进行账号注册，注册的验证码是：%s，十分钟内有效", c),
		fmt.Sprintf(userconstant.RegisterCodeKey, req.Email),
		c,
	)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetRegisterSMSCode(ctx context.Context, req dto.GetSMSCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	c, err := smsx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.sms.SendCode(ctx, fmt.Sprintf(userconstant.RegisterSMSCodeKey, phone), phone, c)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetResetCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c, err := emailx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.email.SendCode(
		ctx,
		req.Email,
		"重置密码",
		fmt.Sprintf("你正在进行账号密码重置，重置的验证码是：%s，十分钟内有效", c),
		fmt.Sprintf(userconstant.ResetPwdCodeKey, req.Email),
		c,
	)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) UpdateProfile(ctx context.Context, userID string, req dto.UpdateUserAdditionReq) (dto.UpdateUserAdditionResp, error) {
	var resp dto.UpdateUserAdditionResp
	if req.Gender != nil && *req.Gender != "" {
		if *req.Gender != "male" && *req.Gender != "female" {
			return resp, ErrInvalidGender
		}
		// 事务统一更新性别，保证两表一致
		if err := ul.userRepo.UpdateGender(ctx, userID, *req.Gender); err != nil {
			if errors.Is(err, repo.ErrUserNotExist) {
				return resp, ErrUserNotExist
			}
			if errors.Is(err, repo.ErrUserUpdateFailed) {
				return resp, ErrProfileUpdateFailed
			}
			ul.log.Error(err)
			return resp, ErrDefault
		}
	}
	var phone *string
	if req.Phone != nil && *req.Phone != "" {
		p := strings.TrimSpace(*req.Phone)
		if !isValidLoosePhone(p) {
			return resp, ErrInvalidPhone
		}
		phone = &p
	}
	var birthday *int64
	if req.Birthday != nil && *req.Birthday != "" {
		t, err := time.Parse("20060102", *req.Birthday)
		if err != nil {
			return resp, ErrInvalidBirthdayFormat
		}
		ms := t.UnixMilli()
		birthday = &ms
	}
	err := ul.userRepo.UpdateAdditionByID(ctx, userID, req.Occupation, phone, req.Province, req.City, nil, birthday)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrUserUpdateFailed) {
			return resp, ErrProfileUpdateFailed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetBindPhoneCode(ctx context.Context, userID string, req dto.GetSMSCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	used, err := ul.userRepo.IsPhoneUsed(ctx, phone, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if used {
		return resp, ErrPhoneIsUsed
	}
	c, err := smsx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.sms.SendCode(ctx, fmt.Sprintf(userconstant.BindPhoneCodeKey, phone), phone, c)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) BindPhone(ctx context.Context, userID string, req dto.BindPhoneReq) (dto.BindContactResp, error) {
	var resp dto.BindContactResp
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	ok, err := ul.sms.VerifyCode(ctx, fmt.Sprintf(userconstant.BindPhoneCodeKey, phone), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	used, err := ul.userRepo.IsPhoneUsed(ctx, phone, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if used {
		return resp, ErrPhoneIsUsed
	}
	err = ul.userRepo.UpdateAdditionByID(ctx, userID, nil, &phone, nil, nil, nil, nil)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrUserUpdateFailed) {
			return resp, ErrProfileUpdateFailed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetBindEmailCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c, err := emailx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.email.SendCode(
		ctx,
		req.Email,
		"绑定邮箱",
		fmt.Sprintf("你正在进行邮箱绑定，验证码是：%s，十分钟内有效", c),
		fmt.Sprintf(userconstant.BindEmailCodeKey, req.Email),
		c,
	)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) BindEmail(ctx context.Context, userID string, req dto.BindEmailReq) (dto.BindContactResp, error) {
	var resp dto.BindContactResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(userconstant.BindEmailCodeKey, req.Email), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	err = ul.userRepo.BindEmail(ctx, userID, req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrEmailIsUsed) {
			return resp, ErrEmailIsUsed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetRebindPhoneCode(ctx context.Context, userID string, req dto.GetSMSCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	used, err := ul.userRepo.IsPhoneUsed(ctx, phone, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if used {
		return resp, ErrPhoneIsUsed
	}
	c, err := smsx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.sms.SendCode(ctx, fmt.Sprintf(userconstant.RebindPhoneCodeKey, phone), phone, c)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) RebindPhone(ctx context.Context, userID string, req dto.BindPhoneReq) (dto.BindContactResp, error) {
	var resp dto.BindContactResp
	phone := strings.TrimSpace(req.Phone)
	if !isValidPhone(phone) {
		return resp, ErrInvalidPhone
	}
	ok, err := ul.sms.VerifyCode(ctx, fmt.Sprintf(userconstant.RebindPhoneCodeKey, phone), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	used, err := ul.userRepo.IsPhoneUsed(ctx, phone, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if used {
		return resp, ErrPhoneIsUsed
	}
	err = ul.userRepo.UpdateAdditionByID(ctx, userID, nil, &phone, nil, nil, nil, nil)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrUserUpdateFailed) {
			return resp, ErrProfileUpdateFailed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) GetRebindEmailCode(ctx context.Context, userID string, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c, err := emailx.GenCode()
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	err = ul.email.SendCode(
		ctx,
		req.Email,
		"换绑邮箱",
		fmt.Sprintf("你正在进行邮箱换绑，验证码是：%s，十分钟内有效", c),
		fmt.Sprintf(userconstant.RebindEmailCodeKey, req.Email),
		c,
	)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) RebindEmail(ctx context.Context, userID string, req dto.BindEmailReq) (dto.BindContactResp, error) {
	var resp dto.BindContactResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(userconstant.RebindEmailCodeKey, req.Email), req.Code)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	err = ul.userRepo.BindEmail(ctx, userID, req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrEmailIsUsed) {
			return resp, ErrEmailIsUsed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

var loosePhoneRe = regexp.MustCompile(`^\+?\d{6,20}$`)
var cnPhoneRe = regexp.MustCompile(`^(?:\+?86)?1[3-9]\d{9}$`)

func isValidLoosePhone(phone string) bool {
	p := strings.TrimSpace(phone)
	return loosePhoneRe.MatchString(p)
}

func isValidPhone(phone string) bool {
	p := strings.TrimSpace(phone)
	return cnPhoneRe.MatchString(p)
}

func (ul *UserLogic) UpdateAvatar(ctx context.Context, userID string, req dto.UpdateAvatarReq) (dto.UpdateAvatarResp, error) {
	var resp dto.UpdateAvatarResp
	avatar := req.Avatar
	if len(avatar) == 0 {
		return resp, ErrParamsType
	}
	err := ul.userRepo.UpdateAvatarByID(ctx, userID, avatar)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrUserUpdateFailed) {
			return resp, ErrProfileUpdateFailed
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) BindPartner(ctx context.Context, userID string, req dto.PartnerBindReq) (dto.PartnerBindResp, error) {
	var resp dto.PartnerBindResp
	ub, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) || errors.Is(err, repo.ErrAccountOrPwd) {
			return resp, ErrAccountOrPassword
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if ub.UserID == userID {
		return resp, ErrParamsType
	}
	// 性别校验与父母角色确定
	self, err := ul.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if self.Gender == ub.Gender {
		return resp, ErrPartnerGenderMismatch
	}
	// 已绑定校验：若已绑定不同对象则拒绝；若已绑定同一对象则幂等返回
	existingPID, e1 := ul.userRepo.GetPartnerByUserID(ctx, userID)
	if e1 != nil {
		ul.log.Error(e1)
		return resp, ErrDefault
	}
	if existingPID != "" {
		if existingPID == ub.UserID {
			resp.PartnerID = ub.UserID
			profile, _ := ul.userRepo.GetMyProfile(ctx, ub.UserID)
			resp.PartnerUsername = profile.Username
			resp.PartnerAvatar = profile.Avatar
			return resp, nil
		}
		return resp, ErrPartnerAlreadyBound
	}
	fatherID, motherID := userID, ub.UserID
	if self.Gender != "male" { // self female
		fatherID, motherID = ub.UserID, userID
	}
	if err = ul.userRepo.BindPartner(ctx, fatherID, motherID); err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if ul.babySyncer != nil {
		err = ul.babySyncer.SyncPartnerBabies(ctx, fatherID, motherID)
	}
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.PartnerID = ub.UserID
	profile, _ := ul.userRepo.GetMyProfile(ctx, ub.UserID)
	resp.PartnerUsername = profile.Username
	resp.PartnerAvatar = profile.Avatar
	return resp, nil
}

func (ul *UserLogic) GetPartner(ctx context.Context, userID string) (dto.PartnerGetResp, error) {
	var resp dto.PartnerGetResp
	pid, err := ul.userRepo.GetPartnerByUserID(ctx, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.PartnerID = pid
	if pid != "" {
		row, e := ul.userRepo.GetMyProfile(ctx, pid)
		if e != nil {
			if errors.Is(e, repo.ErrUserNotExist) {
				return resp, ErrUserNotExist
			}
			ul.log.Error(e)
			return resp, ErrDefault
		}
		resp.PartnerUsername = row.Username
		resp.PartnerAvatar = row.Avatar
	}
	return resp, nil
}

func (ul *UserLogic) MyProfile(ctx context.Context, userID string) (dto.MyProfileResp, error) {
	var resp dto.MyProfileResp
	row, err := ul.userRepo.GetMyProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.UserID = row.UserID
	resp.Account = row.Account
	resp.Email = row.Email
	resp.Username = row.Username
	resp.Gender = row.Gender
	resp.Avatar = row.Avatar
	resp.Phone = row.Phone
	resp.Occupation = row.Occupation
	resp.Birthday = row.Birthday
	resp.Province = row.Province
	resp.City = row.City
	resp.Ctime = row.Ctime
	resp.Utime = row.Utime
	pid, err := ul.userRepo.GetPartnerByUserID(ctx, userID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.PartnerID = pid
	return resp, nil
}

func (ul *UserLogic) Follow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error) {
	var resp dto.FollowResp
	target := uri.TargetUserID
	if target == "" || target == userID {
		return resp, ErrParamsType
	}
	_, err := ul.userRepo.GetUserByID(ctx, target)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		ul.log.Error(err)
		return resp, ErrDefault
	}
	if err := ul.userRepo.FollowUser(ctx, userID, target); err != nil {
		if errors.Is(err, repo.ErrDefault) {
			ul.log.Error(err)
			return resp, ErrDefault
		}
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) Unfollow(ctx context.Context, userID string, uri dto.FollowReq) (dto.FollowResp, error) {
	var resp dto.FollowResp
	target := uri.TargetUserID
	if target == "" || target == userID {
		return resp, ErrParamsType
	}
	if err := ul.userRepo.UnfollowUser(ctx, userID, target); err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) ListFollowing(ctx context.Context, userID string, req dto.FollowingListReq) (dto.FollowingListResp, error) {
	var resp dto.FollowingListResp
	viewID := userID
	if req.UserID != "" {
		viewID = req.UserID
	}
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.ListFollowing(ctx, viewID, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.FollowingUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.FollowingUserItem{
			UserID:     r.UserID,
			Username:   r.Username,
			Avatar:     r.Avatar,
			FollowTime: r.FollowTime,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}

func (ul *UserLogic) ListFollowers(ctx context.Context, userID string, req dto.FollowersListReq) (dto.FollowersListResp, error) {
	var resp dto.FollowersListResp
	viewID := userID
	if req.UserID != "" {
		viewID = req.UserID
	}
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.ListFollowers(ctx, viewID, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.FollowingUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.FollowingUserItem{
			UserID:     r.UserID,
			Username:   r.Username,
			Avatar:     r.Avatar,
			FollowTime: r.FollowTime,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}

// admin
func (ul *UserLogic) AdminListUsers(ctx context.Context, req dto.AdminListUsersReq) (dto.AdminListUsersResp, error) {
	var resp dto.AdminListUsersResp
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}
	rows, hasMore, err := ul.userRepo.AdminListUsers(ctx, req.Keyword, page, pageSize)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.AdminUserItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.AdminUserItem{
			UserID:   r.UserID,
			Username: r.Username,
			Avatar:   r.Avatar,
		})
	}
	resp.List = items
	resp.HasMore = hasMore
	return resp, nil
}

// admin
func (ul *UserLogic) AdminPromoteToAdmin(ctx context.Context, userID string) (string, error) {
	err := ul.userRepo.AdminUpdateUserRole(ctx, userID, int16(jwtx.ADMIN))
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return "", ErrUserNotExist
		}
		ul.log.Error(err)
		return "", ErrDefault
	}
	return "OK", nil
}
