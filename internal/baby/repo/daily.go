package repo

import (
	"context"
	"nurture/internal/baby/repo/dao"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type FeedingRow struct {
	FeedingID string
	FeedTime  int64
	FeedType  string
	Remark    string
}

type DiaperRow struct {
	DiaperID        string
	ChangeTime      int64
	DiaperType      string
	PeeColor        string
	PoopColor       string
	PoopConsistency string
	Remark          string
}

type DailyStats struct {
	FeedingCount    int64
	SleepDurationMs int64
	DiaperCount     int64
}

func (r *BabyRepo) StartSleep(ctx context.Context, babyID, userID string) (string, int64, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", 0, err
	}
	if err := uid.Scan(userID); err != nil {
		return "", 0, err
	}
	row, err := r.babyDao.StartSleep(ctx, dao.StartSleepParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		r.log.Error(err)
		return "", 0, ErrDefault
	}
	return row.SleepID, row.StartTime, nil
}

func (r *BabyRepo) StopSleep(ctx context.Context, sessionID string) (string, int64, int64, int64, error) {
	var sid pgtype.UUID
	if err := sid.Scan(sessionID); err != nil {
		return "", 0, 0, 0, err
	}
	row, err := r.babyDao.StopSleep(ctx, sid)
	if err != nil {
		r.log.Error(err)
		return "", 0, 0, 0, ErrDefault
	}
	return row.SleepID, row.StartTime, row.EndTime.Int64, row.Duration.Int64, nil
}

func (r *BabyRepo) ForceStopSleepWithCap(ctx context.Context, sessionID string, capMs int64) (string, int64, int64, int64, error) {
	var sid pgtype.UUID
	if err := sid.Scan(sessionID); err != nil {
		return "", 0, 0, 0, err
	}
	row, err := r.babyDao.ForceStopSleepWithCap(ctx, dao.ForceStopSleepWithCapParams{
		SleepID:  sid,
		Duration: pgtype.Int8{Int64: capMs, Valid: true},
	})
	if err != nil {
		r.log.Error(err)
		return "", 0, 0, 0, ErrDefault
	}
	return row.SleepID, row.StartTime, row.EndTime.Int64, row.Duration.Int64, nil
}

func (r *BabyRepo) GetActiveSleep(ctx context.Context, babyID, userID string) (string, int64, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", 0, err
	}
	if err := uid.Scan(userID); err != nil {
		return "", 0, err
	}
	row, err := r.babyDao.GetActiveSleep(ctx, dao.GetActiveSleepParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, nil
		}
		r.log.Error(err)
		return "", 0, ErrDefault
	}
	return row.SleepID, row.StartTime, nil
}

type SleepRow struct {
	SessionID string
	StartTime int64
	EndTime   int64
	Duration  int64
}

func (r *BabyRepo) ListSleepBetween(ctx context.Context, babyID string, from, to int64) ([]SleepRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListSleepByBabyBetween(ctx, dao.ListSleepByBabyBetweenParams{
		BabyID:    bid,
		EndTime:   pgtype.Int8{Int64: from, Valid: true},
		StartTime: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]SleepRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, SleepRow{
			SessionID: v.SessionID,
			StartTime: v.StartTime,
			EndTime:   v.EndTime.Int64,
			Duration:  v.Duration.Int64,
		})
	}
	return items, nil
}

func (r *BabyRepo) CreateFeeding(ctx context.Context, babyID, userID string, feedTime int64, feedType, remark string, now int64) (string, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", err
	}
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := r.babyDao.CreateFeeding(ctx, dao.CreateFeedingParams{
		BabyID:   bid,
		UserID:   uid,
		FeedTime: feedTime,
		FeedType: feedType,
		Column5:  remark,
		Ctime:    now,
	})
	if err != nil {
		r.log.Error(err)
		return "", ErrDefault
	}
	return row.FeedingID, nil
}

func (r *BabyRepo) UpdateFeeding(ctx context.Context, babyID, feedingID string, feedType string, feedTime int64, remark string, now int64) error {
	var bid, fid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return err
	}
	if err := fid.Scan(feedingID); err != nil {
		return err
	}
	_, err := r.babyDao.UpdateFeeding(ctx, dao.UpdateFeedingParams{
		BabyID:    bid,
		FeedingID: fid,
		FeedType:  feedType,
		FeedTime:  feedTime,
		Column5:   remark,
		Utime:     now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *BabyRepo) ListFeedingBetween(ctx context.Context, babyID string, from, to int64) ([]FeedingRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListFeedingByBabyBetween(ctx, dao.ListFeedingByBabyBetweenParams{
		BabyID:     bid,
		FeedTime:   from,
		FeedTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]FeedingRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, FeedingRow{
			FeedingID: v.FeedingID,
			FeedTime:  v.FeedTime,
			FeedType:  v.FeedType,
			Remark:    v.Remark,
		})
	}
	return items, nil
}

func (r *BabyRepo) CreateDiaper(ctx context.Context, babyID, userID string, changeTime int64, diaperType, peeColor, poopColor, poopConsistency, remark string, now int64) (string, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", err
	}
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := r.babyDao.CreateDiaper(ctx, dao.CreateDiaperParams{
		BabyID:     bid,
		UserID:     uid,
		ChangeTime: changeTime,
		DiaperType: diaperType,
		Column5:    peeColor,
		Column6:    poopColor,
		Column7:    poopConsistency,
		Column8:    remark,
		Ctime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return "", ErrDefault
	}
	return row.DiaperID, nil
}

func (r *BabyRepo) UpdateDiaper(ctx context.Context, babyID, diaperID string, diaperType string, changeTime int64, peeColor, poopColor, poopConsistency, remark string, now int64) error {
	var bid, did pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return err
	}
	if err := did.Scan(diaperID); err != nil {
		return err
	}
	_, err := r.babyDao.UpdateDiaper(ctx, dao.UpdateDiaperParams{
		BabyID:     bid,
		DiaperID:   did,
		DiaperType: diaperType,
		ChangeTime: changeTime,
		Column5:    peeColor,
		Column6:    poopColor,
		Column7:    poopConsistency,
		Column8:    remark,
		Utime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *BabyRepo) GetDiaperBetween(ctx context.Context, babyID string, from, to int64) (DiaperRow, bool, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return DiaperRow{}, false, err
	}
	row, err := r.babyDao.GetDiaperByBabyBetween(ctx, dao.GetDiaperByBabyBetweenParams{
		BabyID:       bid,
		ChangeTime:   from,
		ChangeTime_2: to,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiaperRow{}, false, nil
		}
		r.log.Error(err)
		return DiaperRow{}, false, ErrDefault
	}
	return DiaperRow{
		DiaperID:        row.DiaperID,
		ChangeTime:      row.ChangeTime,
		DiaperType:      row.DiaperType,
		PeeColor:        row.PeeColor,
		PoopColor:       row.PoopColor,
		PoopConsistency: row.PoopConsistency,
		Remark:          row.Remark,
	}, true, nil
}

func (r *BabyRepo) ListDiaperBetween(ctx context.Context, babyID string, from, to int64) ([]DiaperRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListDiaperByBabyBetween(ctx, dao.ListDiaperByBabyBetweenParams{
		BabyID:       bid,
		ChangeTime:   from,
		ChangeTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]DiaperRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, DiaperRow{
			DiaperID:        v.DiaperID,
			ChangeTime:      v.ChangeTime,
			DiaperType:      v.DiaperType,
			PeeColor:        v.PeeColor,
			PoopColor:       v.PoopColor,
			PoopConsistency: v.PoopConsistency,
			Remark:          v.Remark,
		})
	}
	return items, nil
}

func (r *BabyRepo) GetDailyStats(ctx context.Context, babyID string, from, to int64) (DailyStats, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return DailyStats{}, err
	}
	row, err := r.babyDao.GetDailyStatsByBaby(ctx, dao.GetDailyStatsByBabyParams{
		BabyID:     bid,
		FeedTime:   from,
		FeedTime_2: to,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DailyStats{}, nil
		}
		r.log.Error(err)
		return DailyStats{}, ErrDefault
	}
	return DailyStats{
		FeedingCount:    row.FeedingCount,
		SleepDurationMs: row.SleepDurationMs,
		DiaperCount:     row.DiaperCount,
	}, nil
}
