package test

import (
	"context"
	"errors"
	"testing"

	baby "nurture/internal/baby"
	babyrepo "nurture/internal/baby/repo"
	babydao "nurture/internal/baby/repo/dao"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestBabyClientReturnsBoundaryTypesForGrowthReads(t *testing.T) {
	babyID := "33333333-3333-3333-3333-333333333333"
	userID := "11111111-1111-1111-1111-111111111111"
	repo := &babyClientRepoFake{
		baby: babydao.Baby{
			BabyID:   mustUUID(t, babyID),
			Name:     "Nora",
			Gender:   "female",
			Birthday: 1710000000000,
			Avatar:   "avatar.png",
		},
		growthRows: []babydao.BabyGrowthRecord{
			{
				RecordTime:        1710000001000,
				Height:            pgtype.Float8{Float64: 65.5, Valid: true},
				Weight:            pgtype.Float8{Float64: 7.2, Valid: true},
				HeadCircumference: pgtype.Float8{Float64: 42.1, Valid: true},
				Remark:            pgtype.Text{String: "steady", Valid: true},
			},
			{
				RecordTime: 1710000002000,
			},
		},
	}
	client := baby.NewClient(repo)

	profile, err := client.GetBabyByIDAndUser(context.Background(), babyID, userID)
	if err != nil {
		t.Fatalf("GetBabyByIDAndUser() error = %v", err)
	}
	if profile.BabyID != babyID || profile.Name != "Nora" || profile.Gender != "female" || profile.Birthday != 1710000000000 || profile.Avatar != "avatar.png" {
		t.Fatalf("GetBabyByIDAndUser() = %+v, want converted baby profile", profile)
	}

	records, err := client.ListGrowthRecordsByBabyIDBetween(context.Background(), babyID, 1, 2)
	if err != nil {
		t.Fatalf("ListGrowthRecordsByBabyIDBetween() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("ListGrowthRecordsByBabyIDBetween() len = %d, want 2", len(records))
	}
	first := records[0]
	if first.Height == nil || *first.Height != 65.5 {
		t.Fatalf("first.Height = %v, want 65.5", first.Height)
	}
	if first.Weight == nil || *first.Weight != 7.2 {
		t.Fatalf("first.Weight = %v, want 7.2", first.Weight)
	}
	if first.HeadCircumference == nil || *first.HeadCircumference != 42.1 {
		t.Fatalf("first.HeadCircumference = %v, want 42.1", first.HeadCircumference)
	}
	if first.Remark != "steady" {
		t.Fatalf("first.Remark = %q, want steady", first.Remark)
	}
	if records[1].Height != nil || records[1].Weight != nil || records[1].HeadCircumference != nil || records[1].Remark != "" {
		t.Fatalf("second record = %+v, want nil optional metrics and empty remark", records[1])
	}
}

func TestBabyClientMapsMissingBabyToModuleError(t *testing.T) {
	client := baby.NewClient(&babyClientRepoFake{babyErr: babyrepo.ErrBabyNotExist})

	_, err := client.GetBabyByIDAndUser(context.Background(), "baby-1", "user-1")
	if !errors.Is(err, baby.ErrBabyNotExist) {
		t.Fatalf("GetBabyByIDAndUser() error = %v, want %v", err, baby.ErrBabyNotExist)
	}
}

func TestBabyModuleExposesClient(t *testing.T) {
	module := baby.NewModule(baby.Deps{})
	if module.Client() == nil {
		t.Fatal("Client() = nil, want non-nil")
	}
}

type babyClientRepoFake struct {
	baby       babydao.Baby
	babyErr    error
	growthRows []babydao.BabyGrowthRecord
	growthErr  error
}

func (f *babyClientRepoFake) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (babydao.Baby, error) {
	return f.baby, f.babyErr
}

func (f *babyClientRepoFake) ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, from, to int64) ([]babydao.BabyGrowthRecord, error) {
	return f.growthRows, f.growthErr
}

func mustUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("scan uuid %q: %v", value, err)
	}
	return id
}
