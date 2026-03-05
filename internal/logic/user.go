package logic

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/constant"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/repo"
	"time"

	"github.com/google/uuid"
)

type IUserLogic interface {
	Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
	GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetRegisterCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetResetCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	ResetPassword(ctx context.Context, req dto.ResetPasswordReq) (dto.ResetPasswordResp, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateUserAdditionReq) (dto.UpdateUserAdditionResp, error)
	UpdateAvatar(ctx context.Context, userID string, req dto.UpdateAvatarReq) (dto.UpdateAvatarResp, error)
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
	userRepo *repo.UserRepo
	email    *emailx.EmailX
}

func NewUserLogic() *UserLogic {
	return &UserLogic{
		userRepo: repo.NewUserRepo(),
		email:    emailx.NewEmailX(),
	}
}

var _ IUserLogic = (*UserLogic)(nil)

func (ul *UserLogic) Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error) {
	var resp dto.LoginResp
	switch req.LoginType {
	case constant.LOGIN_WITH_ACCOUNT:
		data, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
		if err != nil {
			return resp, ErrAccountOrPassword
		}
		token, err := jwtx.GenToken(jwtx.Claims{
			UserID: data.UserID.String(),
			Role:   jwtx.Role(data.Role),
		})
		if err != nil {
			global.Log.Error(err)
			return resp, ErrDefault
		}
		resp.Token = token
		return resp, nil
	case constant.LOGIN_WITH_EMAIL:
		ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(constant.LOGIN_CODE_KEY, req.Email), req.Code)
		if err != nil {
			global.Log.Error(err)
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
			UserID: data.UserID.String(),
			Role:   jwtx.Role(data.Role),
		})
		if err != nil {
			global.Log.Error(err)
			return resp, ErrDefault
		}
		resp.Token = token
		return resp, nil
	default:
		global.Log.Warnf("错误的登录方式:%s", req.LoginType)
		return resp, ErrLoginWithFailedWay
	}
}

func (ul *UserLogic) Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error) {
	var resp dto.RegisterResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(constant.REGISTER_CODE_KEY, req.Email), req.Code)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	err = ul.userRepo.Register(ctx, uuid.NewString(), req.Username, req.Email, req.Account, req.Password, req.Gender)
	if err != nil {
		if errors.Is(err, repo.ErrEmailIsUsed) {
			return resp, ErrEmailIsUsed
		} else if errors.Is(err, repo.ErrAccountIsUsed) {
			return resp, ErrAccountIsUsed
		} else {
			global.Log.Error(err)
			return resp, ErrDefault
		}
	}
	resp.Message = "用户注册成功！"
	return resp, nil
}

func (ul *UserLogic) ResetPassword(ctx context.Context, req dto.ResetPasswordReq) (dto.ResetPasswordResp, error) {
	var resp dto.ResetPasswordResp
	ok, err := ul.email.VerifyCode(ctx, fmt.Sprintf(constant.RESET_PWD_CODE_KEY, req.Email), req.Code)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrCodeVerify
	}
	if !ok {
		return resp, ErrCodeVerify
	}
	err = ul.userRepo.ResetPassword(ctx, req.Email, req.NewPassword)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "重置密码成功！"
	return resp, nil
}

func (ul *UserLogic) GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c := emailx.GenCode()
	err := ul.email.SendLoginCode(ctx, req.Email, c)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Code = c
	return resp, nil
}

func (ul *UserLogic) GetRegisterCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c := emailx.GenCode()
	err := ul.email.SendRegisterCode(ctx, req.Email, c)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Code = c
	return resp, nil
}

func (ul *UserLogic) GetResetCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error) {
	var resp dto.GetCodeResp
	c := emailx.GenCode()
	err := ul.email.SendResetPwdCode(ctx, req.Email, c)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrCodeGet
	}
	resp.Code = c
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
			global.Log.Error(err)
			return resp, ErrDefault
		}
	}
	if req.Phone != nil && *req.Phone != "" {
		// 简单手机号正则：可选+，6-20位数字
		phone := *req.Phone
		valid := false
		for i := 0; i < len(phone); i++ {
			c := phone[i]
			if i == 0 && c == '+' {
				continue
			}
			if c < '0' || c > '9' {
				valid = false
				break
			}
			valid = true
		}
		if !valid || len(phone) < 6 || len(phone) > 20 {
			return resp, ErrInvalidPhone
		}
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
	err := ul.userRepo.UpdateAdditionByID(ctx, userID, req.Occupation, req.Phone, req.Province, req.City, nil, birthday)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		if errors.Is(err, repo.ErrUserUpdateFailed) {
			return resp, ErrProfileUpdateFailed
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "OK"
	return resp, nil
}

func (ul *UserLogic) BindPartner(ctx context.Context, userID string, req dto.PartnerBindReq) (dto.PartnerBindResp, error) {
	var resp dto.PartnerBindResp
	ub, err := ul.userRepo.LoginWithAccount(ctx, req.Account, req.Password)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		global.Log.Error(err)
		return resp, ErrAccountOrPassword
	}
	if ub.UserID.String() == userID {
		return resp, ErrParamsType
	}
	// 性别校验与父母角色确定
	self, err := ul.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotExist) {
			return resp, ErrUserNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if self.Gender == ub.Gender {
		return resp, ErrPartnerGenderMismatch
	}
	// 已绑定校验：若已绑定不同对象则拒绝；若已绑定同一对象则幂等返回
	existingPID, e1 := ul.userRepo.GetPartnerByUserID(ctx, userID)
	if e1 != nil {
		global.Log.Error(e1)
		return resp, ErrDefault
	}
	if existingPID != "" {
		if existingPID == ub.UserID.String() {
			resp.PartnerID = ub.UserID.String()
			profile, _ := ul.userRepo.GetMyProfile(ctx, ub.UserID.String())
			resp.PartnerUsername = profile.Username
			resp.PartnerAvatar = profile.Avatar
			return resp, nil
		}
		return resp, ErrPartnerAlreadyBound
	}
	fatherID, motherID := userID, ub.UserID.String()
	if self.Gender != "male" { // self female
		fatherID, motherID = ub.UserID.String(), userID
	}
	err = ul.userRepo.BindPartnerAndSyncBabies(ctx, fatherID, motherID)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.PartnerID = ub.UserID.String()
	profile, _ := ul.userRepo.GetMyProfile(ctx, ub.UserID.String())
	resp.PartnerUsername = profile.Username
	resp.PartnerAvatar = profile.Avatar
	return resp, nil
}

func (ul *UserLogic) GetPartner(ctx context.Context, userID string) (dto.PartnerGetResp, error) {
	var resp dto.PartnerGetResp
	pid, err := ul.userRepo.GetPartnerByUserID(ctx, userID)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.PartnerID = pid
	if pid != "" {
		row, e := ul.userRepo.GetMyProfile(ctx, pid)
		if e != nil {
			if errors.Is(e, repo.ErrUserNotExist) {
				return resp, ErrUserNotExist
			}
			global.Log.Error(e)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if err := ul.userRepo.FollowUser(ctx, userID, target); err != nil {
		if errors.Is(err, repo.ErrDefault) {
			global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
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
		global.Log.Error(err)
		return "", ErrDefault
	}
	return "OK", nil
}
