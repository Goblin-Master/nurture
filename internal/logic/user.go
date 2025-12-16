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

	"github.com/google/uuid"
)

type IUserLogic interface {
	Login(ctx context.Context, req dto.LoginReq) (dto.LoginResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error)
	GetLoginCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetRegisterCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	GetResetCode(ctx context.Context, req dto.GetCodeReq) (dto.GetCodeResp, error)
	ResetPassword(ctx context.Context, req dto.ResetPasswordReq) (dto.ResetPasswordResp, error)
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
		resp.Username = data.Username
		resp.Avatar = data.Avatar
		resp.Token = token
		return resp, nil
	case constant.LOGIN_WITH_EMAIL:
		if ok := ul.email.VerifyCode(fmt.Sprintf(constant.LOGIN_CODE_KEY, req.Email), req.Code); !ok {
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
		resp.Username = data.Username
		resp.Avatar = data.Avatar
		resp.Token = token
		return resp, nil
	default:
		global.Log.Warnf("错误的登录方式:%s", req.LoginType)
		return resp, ErrLoginWithFailedWay
	}
}

func (ul *UserLogic) Register(ctx context.Context, req dto.RegisterReq) (dto.RegisterResp, error) {
	var resp dto.RegisterResp
	if ok := ul.email.VerifyCode(fmt.Sprintf(constant.REGISTER_CODE_KEY, req.Email), req.Code); !ok {
		return resp, ErrCodeVerify
	}
	err := ul.userRepo.Register(ctx, uuid.NewString(), req.Username, req.Email, req.Account, req.Password)
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
	if ok := ul.email.VerifyCode(fmt.Sprintf(constant.RESET_PWD_CODE_KEY, req.Email), req.Code); !ok {
		return resp, ErrCodeVerify
	}
	err := ul.userRepo.ResetPassword(ctx, req.Email, req.NewPassword)
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
