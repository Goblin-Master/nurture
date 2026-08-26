package baby

import (
	"context"
	"errors"

	"nurture/internal/baby/repo"
	"nurture/internal/baby/repo/dao"
)

type BabyProfile struct {
	BabyID   string
	Name     string
	Gender   string
	Birthday int64
	Avatar   string
}

type GrowthRecord struct {
	RecordTime        int64
	Height            *float64
	Weight            *float64
	HeadCircumference *float64
	Remark            string
}

type profileReader interface {
	GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (dao.Baby, error)
	ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, start, end int64) ([]dao.BabyGrowthRecord, error)
}

type Client struct {
	profileReader profileReader
}

func NewClient(profileReader profileReader) *Client {
	return &Client{profileReader: profileReader}
}

func (c *Client) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (BabyProfile, error) {
	b, err := c.profileReader.GetBabyByIDAndUser(ctx, babyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return BabyProfile{}, ErrBabyNotExist
		}
		return BabyProfile{}, err
	}
	return BabyProfile{
		BabyID:   b.BabyID.String(),
		Name:     b.Name,
		Gender:   b.Gender,
		Birthday: b.Birthday,
		Avatar:   b.Avatar,
	}, nil
}

func (c *Client) ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, start, end int64) ([]GrowthRecord, error) {
	rows, err := c.profileReader.ListGrowthRecordsByBabyIDBetween(ctx, babyID, start, end)
	if err != nil {
		return nil, err
	}
	ret := make([]GrowthRecord, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, growthRecordFromDAO(row))
	}
	return ret, nil
}

func growthRecordFromDAO(row dao.BabyGrowthRecord) GrowthRecord {
	item := GrowthRecord{RecordTime: row.RecordTime}
	if row.Height.Valid {
		item.Height = ptrFloat64(row.Height.Float64)
	}
	if row.Weight.Valid {
		item.Weight = ptrFloat64(row.Weight.Float64)
	}
	if row.HeadCircumference.Valid {
		item.HeadCircumference = ptrFloat64(row.HeadCircumference.Float64)
	}
	if row.Remark.Valid {
		item.Remark = row.Remark.String
	}
	return item
}

func ptrFloat64(v float64) *float64 {
	return &v
}
