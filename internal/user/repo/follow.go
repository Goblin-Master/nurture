package repo

import (
	"context"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/repo/cache"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type FollowUserRow struct {
	UserID     string
	Username   string
	Avatar     string
	FollowTime int64
}

func toFollowUserRow(userID, username, avatar string, followTime int64) FollowUserRow {
	return FollowUserRow{
		UserID:     userID,
		Username:   username,
		Avatar:     avatar,
		FollowTime: followTime,
	}
}

func (ur *UserRepo) FollowUser(ctx context.Context, followerID, followeeID string) error {
	var fUID, eUID pgtype.UUID
	if err := fUID.Scan(followerID); err != nil {
		return err
	}
	if err := eUID.Scan(followeeID); err != nil {
		return err
	}
	err := ur.userDao.CreateFollow(ctx, dao.CreateFollowParams{
		Follower: fUID,
		Followee: eUID,
		Ctime:    time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(followerID), cache.ProfileKey(followeeID))
	_ = cache.ScanDel(ctx, ur.rdb, cache.FollowingPattern(followerID), 100)
	_ = cache.ScanDel(ctx, ur.rdb, cache.FollowersPattern(followeeID), 100)
	return nil
}

func (ur *UserRepo) UnfollowUser(ctx context.Context, followerID, followeeID string) error {
	var fUID, eUID pgtype.UUID
	if err := fUID.Scan(followerID); err != nil {
		return err
	}
	if err := eUID.Scan(followeeID); err != nil {
		return err
	}
	n, err := ur.userDao.DeleteFollow(ctx, dao.DeleteFollowParams{
		Follower: fUID,
		Followee: eUID,
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	if n == 0 {
		return nil
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(followerID), cache.ProfileKey(followeeID))
	_ = cache.ScanDel(ctx, ur.rdb, cache.FollowingPattern(followerID), 100)
	_ = cache.ScanDel(ctx, ur.rdb, cache.FollowersPattern(followeeID), 100)
	return nil
}

func (ur *UserRepo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var fUID, eUID pgtype.UUID
	if err := fUID.Scan(followerID); err != nil {
		return false, err
	}
	if err := eUID.Scan(followeeID); err != nil {
		return false, err
	}
	ok, err := ur.userDao.IsFollowing(ctx, dao.IsFollowingParams{
		Follower: fUID,
		Followee: eUID,
	})
	if err != nil {
		ur.log.Error(err)
		return false, ErrDefault
	}
	return ok, nil
}

func (ur *UserRepo) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]FollowUserRow, bool, error) {
	type listCache struct {
		Rows    []FollowUserRow `json:"rows"`
		HasMore bool            `json:"has_more"`
	}
	key := cache.FollowingKey(userID, page, pageSize)
	{
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, ur.rdb, key, &cached); ok {
			ur.logCacheHit("user_cache_hit", "type", "following", "user", userID, "page", page, "size", pageSize)
			return cached.Rows, cached.HasMore, nil
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, err
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.ListFollowingByUserID(ctx, dao.ListFollowingByUserIDParams{
		Follower: uid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		ur.log.Error(err)
		return nil, false, ErrDefault
	}
	_ = cache.Del(ctx, ur.rdb, key)
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	ret := make([]FollowUserRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, toFollowUserRow(row.UserID, row.Username, row.Avatar, row.FollowTime))
	}
	payload := listCache{Rows: ret, HasMore: hasMore}
	_ = cache.SetJSON(ctx, ur.rdb, key, payload, time.Duration(userconstant.ListTTL)*time.Second)
	return ret, hasMore, nil
}

func (ur *UserRepo) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]FollowUserRow, bool, error) {
	type listCache struct {
		Rows    []FollowUserRow `json:"rows"`
		HasMore bool            `json:"has_more"`
	}
	key := cache.FollowersKey(userID, page, pageSize)
	{
		var cached listCache
		if ok, _ := cache.GetJSON(ctx, ur.rdb, key, &cached); ok {
			ur.logCacheHit("user_cache_hit", "type", "followers", "user", userID, "page", page, "size", pageSize)
			return cached.Rows, cached.HasMore, nil
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, err
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.ListFollowersByUserID(ctx, dao.ListFollowersByUserIDParams{
		Followee: uid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		ur.log.Error(err)
		return nil, false, ErrDefault
	}
	_ = cache.Del(ctx, ur.rdb, key)
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	ret := make([]FollowUserRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, toFollowUserRow(row.UserID, row.Username, row.Avatar, row.FollowTime))
	}
	payload := listCache{Rows: ret, HasMore: hasMore}
	_ = cache.SetJSON(ctx, ur.rdb, key, payload, time.Duration(userconstant.ListTTL)*time.Second)
	return ret, hasMore, nil
}
