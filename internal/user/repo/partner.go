package repo

import (
	"context"
	"errors"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/repo/cache"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetPartnerByUserID 查询另一半（返回对方ID；没有则返回空字符串）
func (ur *UserRepo) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	key := cache.PartnerKey(userID)
	if s, ok, _ := cache.GetString(ctx, ur.rdb, key); ok {
		ur.logCacheHit("user_cache_hit", "type", "partner", "user", userID)
		return s, nil
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := ur.userDao.GetPartnerByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = cache.SetEX(ctx, ur.rdb, key, "", time.Duration(userconstant.ProfileTTL)*time.Second)
			return "", nil
		}
		ur.log.Error(err)
		return "", ErrDefault
	}
	if row.Father == userID {
		_ = cache.SetEX(ctx, ur.rdb, key, row.Mother, time.Duration(userconstant.PartnerTTL)*time.Second)
		return row.Mother, nil
	}
	_ = cache.SetEX(ctx, ur.rdb, key, row.Father, time.Duration(userconstant.PartnerTTL)*time.Second)
	return row.Father, nil
}

func (ur *UserRepo) BindPartner(ctx context.Context, fatherUserID, motherUserID string) error {
	var aUUID, bUUID pgtype.UUID
	if scanErr := aUUID.Scan(fatherUserID); scanErr != nil {
		return scanErr
	}
	if scanErr := bUUID.Scan(motherUserID); scanErr != nil {
		return scanErr
	}
	if err := ur.userDao.CreatePartner(ctx, dao.CreatePartnerParams{
		Father: aUUID,
		Mother: bUUID,
		Ctime:  time.Now().UnixMilli(),
	}); err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	_ = cache.Del(ctx, ur.rdb, cache.PartnerKey(fatherUserID), cache.PartnerKey(motherUserID))
	for _, uid := range []string{fatherUserID, motherUserID} {
		_ = cache.ScanDel(ctx, ur.rdb, cache.FollowingPattern(uid), 100)
		_ = cache.ScanDel(ctx, ur.rdb, cache.FollowersPattern(uid), 100)
		_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(uid))
	}
	return nil
}
