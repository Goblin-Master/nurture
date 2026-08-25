package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"time"

	"github.com/google/uuid"
)

func (l *BabyLogic) CreateGrowth(ctx context.Context, userID string, req dto.CreateGrowthReq) (dto.CreateGrowthResp, error) {
	var resp dto.CreateGrowthResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	h := req.Height
	w := req.Weight
	hc := req.HeadCircumference
	gr, e := l.babyRepo.GetLatestGrowthByBabyID(ctx, req.BabyID)
	if e != nil && !errors.Is(e, repo.ErrBabyGrowthNotExist) {
		l.log.Error(e)
		return resp, ErrDefault
	}
	if e == nil {
		if h == 0 && gr.Height.Valid {
			h = gr.Height.Float64
		}
		if w == 0 && gr.Weight.Valid {
			w = gr.Weight.Float64
		}
		if hc == 0 && gr.HeadCircumference.Valid {
			hc = gr.HeadCircumference.Float64
		}
		rtDay := time.UnixMilli(req.RecordTime).UTC().Truncate(24 * time.Hour).UnixMilli()
		grDay := time.UnixMilli(gr.RecordTime).UTC().Truncate(24 * time.Hour).UnixMilli()
		if rtDay == grDay {
			if err := l.babyRepo.UpdateGrowthByRecordID(ctx, gr.RecordID.String(), req.RecordTime, h, w, hc, req.Remark, userID); err != nil {
				l.log.Error(err)
				return resp, ErrDefault
			}
			resp.RecordID = gr.RecordID.String()
			resp.Message = "更新成功"
			return resp, nil
		}
	}
	rid, err := l.babyRepo.CreateGrowthRecord(ctx, req.BabyID, userID, req.RecordTime, h, w, hc, req.Remark)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.RecordID = rid
	resp.Message = "创建成功"
	return resp, nil
}

func (l *BabyLogic) GetGrowthAt(ctx context.Context, userID string, req dto.GrowthAtReq) (dto.GrowthAtResp, error) {
	var resp dto.GrowthAtResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	dayStart := time.UnixMilli(req.Time).UTC().Truncate(24 * time.Hour)
	start := dayStart.UnixMilli()
	end := dayStart.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	gr, e := l.babyRepo.GetGrowthByBabyIDBetween(ctx, req.BabyID, start, end)
	if e != nil {
		if errors.Is(e, repo.ErrBabyGrowthNotExist) {
			return resp, nil
		}
		l.log.Error(e)
		return resp, ErrDefault
	}
	resp.RecordID = gr.RecordID.String()
	resp.RecordTime = gr.RecordTime
	if gr.Height.Valid {
		v := gr.Height.Float64
		resp.Height = &v
	}
	if gr.Weight.Valid {
		v := gr.Weight.Float64
		resp.Weight = &v
	}
	if gr.HeadCircumference.Valid {
		v := gr.HeadCircumference.Float64
		resp.HeadCircumference = &v
	}
	if gr.Remark.Valid {
		resp.Remark = gr.Remark.String
	}
	if gr.CreatedBy.Valid {
		if uid, err := uuid.FromBytes(gr.CreatedBy.Bytes[:]); err == nil {
			resp.CreatedBy = uid.String()
		}
	}
	if gr.UpdatedBy.Valid {
		if uid, err := uuid.FromBytes(gr.UpdatedBy.Bytes[:]); err == nil {
			resp.UpdatedBy = uid.String()
		}
	}
	return resp, nil
}

func (l *BabyLogic) GrowthCurve(ctx context.Context, userID string, req dto.GrowthCurveReq) (dto.GrowthCurveResp, error) {
	var resp dto.GrowthCurveResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	to := req.To
	if to <= 0 || to > now {
		to = now
	}
	from := req.From
	if from > to {
		return resp, ErrParamsType
	}
	var unit string
	var rows []repo.CurvePoint
	switch req.Metric {
	case "height":
		unit = "cm"
		rows, err = l.babyRepo.ListHeightCurveBetween(ctx, req.BabyID, from, to)
	case "weight":
		unit = "kg"
		rows, err = l.babyRepo.ListWeightCurveBetween(ctx, req.BabyID, from, to)
	case "head_circumference":
		unit = "cm"
		rows, err = l.babyRepo.ListHeadCircumferenceCurveBetween(ctx, req.BabyID, from, to)
	default:
		return resp, ErrParamsType
	}
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	if req.GroupBy == "month" && len(rows) > 0 {
		var grouped []repo.CurvePoint
		curY := -1
		curM := time.Month(0)
		var last repo.CurvePoint
		for _, v := range rows {
			t := time.UnixMilli(v.Time).UTC()
			y, m, _ := t.Date()
			if curY == -1 {
				curY = y
				curM = m
			}
			if y != curY || m != curM {
				if last.Time != 0 {
					grouped = append(grouped, last)
				}
				curY = y
				curM = m
			}
			last = v
		}
		if last.Time != 0 {
			grouped = append(grouped, last)
		}
		rows = grouped
	}
	if req.MaxPoints > 0 && len(rows) > req.MaxPoints {
		n := len(rows)
		stride := (n + req.MaxPoints - 1) / req.MaxPoints
		var sampled []repo.CurvePoint
		for i := 0; i < n; i += stride {
			sampled = append(sampled, rows[i])
		}
		if sampled[len(sampled)-1].Time != rows[n-1].Time {
			sampled = append(sampled, rows[n-1])
		}
		rows = sampled
	}
	items := make([]dto.CurvePoint, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.CurvePoint{
			Time:  v.Time,
			Value: v.Value,
		})
	}
	resp.Metric = req.Metric
	resp.Unit = unit
	resp.Items = items
	return resp, nil
}
