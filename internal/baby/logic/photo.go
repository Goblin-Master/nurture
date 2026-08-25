package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"time"
)

func (l *BabyLogic) UploadBabyPhotos(ctx context.Context, userID string, req dto.UploadBabyPhotosReq) (dto.UploadBabyPhotosResp, error) {
	var resp dto.UploadBabyPhotosResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	rows, err := l.babyRepo.UploadPhotos(ctx, req.BabyID, req.Links, now)
	if err != nil {
		l.log.Error(err)
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
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	n, err := l.babyRepo.DeletePhotos(ctx, req.BabyID, req.PhotoIDs)
	if err != nil {
		l.log.Error(err)
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
		l.log.Error(err)
		return resp, ErrDefault
	}
	rows, hasMore, err := l.babyRepo.ListPhotos(ctx, req.BabyID, req.Page, req.PageSize)
	if err != nil {
		l.log.Error(err)
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
	resp.HasMore = hasMore
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	return resp, nil
}
