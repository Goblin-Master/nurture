package repo

import (
	"context"
	"errors"
	"nurture/internal/post/repo/dao"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostRepo) CreateTag(ctx context.Context, tagID, name, description string, now int64) (TagRow, error) {
	var tid pgtype.UUID
	if err := tid.Scan(tagID); err != nil {
		return TagRow{}, ErrParamsType
	}
	row, err := r.dao.CreateTag(ctx, dao.CreateTagParams{
		TagID:       tid,
		TagName:     name,
		Description: pgtype.Text{String: description, Valid: true},
		Ctime:       now,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return TagRow{}, ErrDefault
		}
		r.log.Error(err)
		return TagRow{}, ErrDefault
	}
	return TagRow{TagID: row.TagID, Name: row.TagName, Description: row.Description}, nil
}

func (r *PostRepo) DeleteTag(ctx context.Context, tagID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.dao.WithTx(tx)
	var tid pgtype.UUID
	if err := tid.Scan(tagID); err != nil {
		_ = tx.Rollback(ctx)
		return ErrParamsType
	}
	if _, err := qtx.DeletePostTagsByTagID(ctx, tid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	if _, err := qtx.DeleteTagByID(ctx, tid); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) ListTags(ctx context.Context, keyword string, page, pageSize int) ([]TagRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := r.dao.ListTags(ctx, dao.ListTagsParams{
		Column1: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		r.log.Error(err)
		return nil, false, ErrDefault
	}
	hasMore := int32(len(rows)) >= limit
	res := make([]TagRow, 0, pageSize)
	for i, v := range rows {
		if int32(i) >= limit-1 {
			break
		}
		res = append(res, TagRow{
			TagID:       v.TagID,
			Name:        v.TagName,
			Description: v.Description,
		})
	}
	return res, hasMore, nil
}
