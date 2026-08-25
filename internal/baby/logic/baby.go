package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"time"

	"github.com/google/uuid"
)

func (l *BabyLogic) ChangeBaby(ctx context.Context, userID string, req dto.ChangeBabyReq) (dto.ChangeBabyResp, error) {
	var resp dto.ChangeBabyResp
	rows, err := l.babyRepo.ListMyBabies(ctx, userID)
	if err != nil {
		return resp, ErrDefault
	}
	items := make([]dto.BabyItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.BabyItem{
			BabyID: v.BabyID,
			Name:   v.Name,
			Avatar: v.Avatar,
		})
	}
	resp.Babies = items
	return resp, nil
}

func (l *BabyLogic) NewBaby(ctx context.Context, userID string, req dto.NewBabyReq) (dto.NewBabyResp, error) {
	var resp dto.NewBabyResp
	babyID := uuid.NewString()
	now := time.Now().UnixMilli()
	var partnerID string
	if l.partnerReader != nil {
		var err error
		partnerID, err = l.partnerReader.GetPartnerByUserID(ctx, userID)
		if err != nil {
			l.log.Error(err)
			return resp, ErrDefault
		}
	}
	err := l.babyRepo.CreateBabyWithInit(ctx, userID, partnerID, babyID, req.Name, req.Gender, req.Birthday, req.Avatar,
		now, req.Height, req.Weight, req.HeadCircumference, req.Remark)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.BabyID = babyID
	resp.Message = "宝宝创建成功"
	return resp, nil
}

func (l *BabyLogic) GetProfile(ctx context.Context, userID string, req dto.BabyProfileReq) (dto.BabyProfileResp, error) {
	var resp dto.BabyProfileResp
	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	gr, err := l.babyRepo.GetLatestGrowthByBabyID(ctx, req.BabyID)
	if err != nil && !errors.Is(err, repo.ErrBabyGrowthNotExist) {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.BabyID = b.BabyID.String()
	resp.Name = b.Name
	resp.Avatar = b.Avatar
	resp.Gender = b.Gender
	resp.Birthday = b.Birthday
	if err == nil {
		resp.RecordTime = gr.RecordTime
		if gr.Height.Valid {
			resp.Height = gr.Height.Float64
		}
		if gr.Weight.Valid {
			resp.Weight = gr.Weight.Float64
		}
		if gr.HeadCircumference.Valid {
			resp.HeadCircumference = gr.HeadCircumference.Float64
		}
	}
	return resp, nil
}
