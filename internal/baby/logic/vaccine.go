package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"time"

	"github.com/google/uuid"
)

func (l *BabyLogic) GetVaccineList(ctx context.Context, userID string, req dto.GetVaccineListReq) (dto.GetVaccineListResp, error) {
	var resp dto.GetVaccineListResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	rows, err := l.babyRepo.ListVaccineRecordsByBaby(ctx, req.BabyID)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.VaccineItem, 0, len(rows))
	for _, v := range rows {
		if req.Status != "" && req.Status != "all" && v.Status != req.Status {
			continue
		}
		items = append(items, dto.VaccineItem{
			DoseID:     v.DoseID,
			VaccineID:  v.VaccineID,
			Name:       v.Name,
			Disease:    v.Disease,
			Link:       v.Link,
			DoseNumber: v.DoseNumber,
			DueTime:    v.DueTime,
			Status:     v.Status,
			ActualTime: v.ActualTime,
		})
	}
	resp.Items = items
	return resp, nil
}

func (l *BabyLogic) AdminCreateVaccine(ctx context.Context, req dto.AdminCreateVaccineReq) (dto.AdminCreateVaccineResp, error) {
	var resp dto.AdminCreateVaccineResp
	if req.Name == "" || req.Disease == "" || len(req.Doses) == 0 {
		return resp, ErrParamsType
	}
	vaccineID := uuid.NewString()
	doses := make([]repo.DoseSpec, 0, len(req.Doses))
	for _, d := range req.Doses {
		if d.DoseNumber <= 0 || d.RecommendAgeDays < 0 {
			return resp, ErrParamsType
		}
		doses = append(doses, repo.DoseSpec{
			DoseNumber:       d.DoseNumber,
			RecommendAgeDays: d.RecommendAgeDays,
		})
	}
	vid, created, err := l.babyRepo.AdminCreateVaccine(ctx, vaccineID, req.Name, req.Disease, req.Link, doses)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.VaccineID = vid
	resp.Doses = make([]dto.AdminCreatedDose, 0, len(created))
	for _, c := range created {
		resp.Doses = append(resp.Doses, dto.AdminCreatedDose{
			DoseID:           c.DoseID,
			DoseNumber:       c.DoseNumber,
			RecommendAgeDays: c.RecommendAgeDays,
		})
	}
	return resp, nil
}

func (l *BabyLogic) ChangeVaccineStatus(ctx context.Context, userID string, req dto.ChangeVaccineStatusReq) (dto.ChangeVaccineStatusResp, error) {
	var resp dto.ChangeVaccineStatusResp
	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	if req.Status == "given" {
		_, err = l.babyRepo.UpdateVaccineStatusGiven(ctx, b.BabyID.String(), req.DoseID, req.ActualTime, now)
	} else {
		_, err = l.babyRepo.UpdateVaccineStatusNotGiven(ctx, b.BabyID.String(), req.DoseID, now)
	}
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "操作成功"
	return resp, nil
}
