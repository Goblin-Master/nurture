package logic

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/passwordx"
	"nurture/internal/pkg/smsx"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
	"strings"

	"github.com/google/uuid"
)

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
