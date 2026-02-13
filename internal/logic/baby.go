package logic

import (
	"context"
	"errors"
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/repo"
	"time"

	"github.com/google/uuid"
)

type IBabyLogic interface {
	ChangeBaby(ctx context.Context, userID string, req dto.ChangeBabyReq) (dto.ChangeBabyResp, error)
	NewBaby(ctx context.Context, userID string, req dto.NewBabyReq) (dto.NewBabyResp, error)
	GetProfile(ctx context.Context, userID string, req dto.BabyProfileReq) (dto.BabyProfileResp, error)
	GetVaccineList(ctx context.Context, userID string, req dto.GetVaccineListReq) (dto.GetVaccineListResp, error)
	ChangeVaccineStatus(ctx context.Context, userID string, req dto.ChangeVaccineStatusReq) (dto.ChangeVaccineStatusResp, error)
	UploadBabyPhotos(ctx context.Context, userID string, req dto.UploadBabyPhotosReq) (dto.UploadBabyPhotosResp, error)
	DeleteBabyPhotos(ctx context.Context, userID string, req dto.DeleteBabyPhotosReq) (dto.DeleteBabyPhotosResp, error)
	ListBabyPhotos(ctx context.Context, userID string, req dto.ListBabyPhotosReq) (dto.ListBabyPhotosResp, error)
}

type BabyLogic struct {
	babyRepo repo.IBabyRepo
}

func NewBabyLogic() *BabyLogic {
	return &BabyLogic{
		babyRepo: repo.NewBabyRepo(),
	}
}

var _ IBabyLogic = (*BabyLogic)(nil)

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
	err := l.babyRepo.CreateBabyWithInit(ctx, userID, babyID, req.Name, req.Gender, req.Birthday, req.Avatar,
		now, req.Height, req.Weight, req.HeadCircumference, req.Remark)
	if err != nil {
		global.Log.Error(err)
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
		global.Log.Error(err)
		return resp, ErrDefault
	}
	gr, err := l.babyRepo.GetLatestGrowthByBabyIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyGrowthNotExist) {
			return resp, ErrBabyGrowthNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.BabyID = b.BabyID.String()
	resp.Name = b.Name
	resp.Avatar = b.Avatar
	resp.Gender = b.Gender
	resp.Birthday = b.Birthday
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
	return resp, nil
}

func (l *BabyLogic) GetVaccineList(ctx context.Context, userID string, req dto.GetVaccineListReq) (dto.GetVaccineListResp, error) {
	var resp dto.GetVaccineListResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	rows, err := l.babyRepo.ListVaccineRecordsByBaby(ctx, req.BabyID)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.VaccineItem, 0, len(rows))
	for _, v := range rows {
		// 过滤状态：status为空或all表示不过滤
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

func (l *BabyLogic) ChangeVaccineStatus(ctx context.Context, userID string, req dto.ChangeVaccineStatusReq) (dto.ChangeVaccineStatusResp, error) {
	var resp dto.ChangeVaccineStatusResp
	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	switch req.Status {
	case "given":
		if req.ActualTime <= 0 || req.ActualTime < b.Birthday || req.ActualTime > now {
			return resp, ErrInvalidActualTime
		}
		n, err := l.babyRepo.UpdateVaccineStatusGiven(ctx, req.BabyID, req.DoseID, req.ActualTime, now)
		if err != nil {
			if errors.Is(err, repo.ErrBabyVaccineNotExist) {
				return resp, ErrVaccineRecordNotExist
			}
			global.Log.Error(err)
			return resp, ErrDefault
		}
		if n == 0 {
			return resp, ErrVaccineRecordNotExist
		}
	case "not_given":
		if req.ActualTime != 0 {
			return resp, ErrInvalidActualTime
		}
		n, err := l.babyRepo.UpdateVaccineStatusNotGiven(ctx, req.BabyID, req.DoseID, now)
		if err != nil {
			if errors.Is(err, repo.ErrBabyVaccineNotExist) {
				return resp, ErrVaccineRecordNotExist
			}
			global.Log.Error(err)
			return resp, ErrDefault
		}
		if n == 0 {
			return resp, ErrVaccineRecordNotExist
		}
	default:
		return resp, ErrInvalidVaccineStatus
	}
	resp.Message = "更新成功"
	return resp, nil
}

func (l *BabyLogic) UploadBabyPhotos(ctx context.Context, userID string, req dto.UploadBabyPhotosReq) (dto.UploadBabyPhotosResp, error) {
	var resp dto.UploadBabyPhotosResp
	if len(req.Links) == 0 {
		return resp, ErrParamsType
	}
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	rows, err := l.babyRepo.UploadPhotos(ctx, req.BabyID, req.Links, now)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.PhotoItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.PhotoItem{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	resp.Items = items
	return resp, nil
}

func (l *BabyLogic) DeleteBabyPhotos(ctx context.Context, userID string, req dto.DeleteBabyPhotosReq) (dto.DeleteBabyPhotosResp, error) {
	var resp dto.DeleteBabyPhotosResp
	if len(req.PhotoIDs) == 0 {
		return resp, ErrParamsType
	}
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	n, err := l.babyRepo.DeletePhotos(ctx, req.BabyID, req.PhotoIDs)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	resp.Deleted = n
	return resp, nil
}

func (l *BabyLogic) ListBabyPhotos(ctx context.Context, userID string, req dto.ListBabyPhotosReq) (dto.ListBabyPhotosResp, error) {
	var resp dto.ListBabyPhotosResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		global.Log.Error(err)
		return resp, ErrDefault
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	rows, hasMore, err := l.babyRepo.ListPhotos(ctx, req.BabyID, req.Page, req.PageSize)
	if err != nil {
		global.Log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.PhotoItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.PhotoItem{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	resp.Items = items
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	resp.HasMore = hasMore
	return resp, nil
}
