package logic

import (
	"context"
	"errors"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
)

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
			profile, err := ul.userRepo.GetMyProfile(ctx, ub.UserID)
			if err != nil {
				ul.log.Error(err)
				return resp, ErrDefault
			}
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
	resp.PartnerID = ub.UserID
	profile, err := ul.userRepo.GetMyProfile(ctx, ub.UserID)
	if err != nil {
		ul.log.Error(err)
		return resp, ErrDefault
	}
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
