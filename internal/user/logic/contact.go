package logic

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/smsx"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
	"regexp"
	"strings"
)

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
