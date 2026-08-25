package repo

import (
	"context"
	"nurture/internal/baby/repo/dao"

	"github.com/jackc/pgx/v5/pgtype"
)

type PhotoRow struct {
	PhotoID string
	Link    string
	Ctime   int64
}

func (r *BabyRepo) UploadPhotos(ctx context.Context, babyID string, links []string, now int64) ([]PhotoRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.UploadBabyPhotos(ctx, dao.UploadBabyPhotosParams{
		BabyID:  bid,
		Column2: links,
		Ctime:   now,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	var items []PhotoRow
	for _, v := range rows {
		items = append(items, PhotoRow{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	return items, nil
}

func (r *BabyRepo) DeletePhotos(ctx context.Context, babyID string, photoIDs []string) (int64, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return 0, err
	}
	ids := make([]pgtype.UUID, 0, len(photoIDs))
	for _, s := range photoIDs {
		var id pgtype.UUID
		if err := id.Scan(s); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	n, err := r.babyDao.DeleteBabyPhotos(ctx, dao.DeleteBabyPhotosParams{
		BabyID:  bid,
		Column2: ids,
	})
	if err != nil {
		r.log.Error(err)
		return 0, ErrDefault
	}
	return n, nil
}

func (r *BabyRepo) ListPhotos(ctx context.Context, babyID string, page, pageSize int) ([]PhotoRow, bool, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, false, err
	}
	offset := (page - 1) * pageSize
	limit := pageSize + 1
	rows, err := r.babyDao.ListBabyPhotos(ctx, dao.ListBabyPhotosParams{
		BabyID: bid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	var items []PhotoRow
	for _, v := range rows {
		items = append(items, PhotoRow{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	hasMore := false
	if len(items) > pageSize {
		hasMore = true
		items = items[:pageSize]
	}
	return items, hasMore, nil
}
