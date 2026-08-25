package logic

import (
	"context"
	"errors"
	"nurture/internal/user/dto"
	"nurture/internal/user/repo"
	"strings"
	"time"
)

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
