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

type ProfileRow struct {
	UserID     string
	Account    string
	Email      string
	Username   string
	Gender     string
	Avatar     string
	Phone      string
	Occupation string
	Birthday   int64
	Province   string
	City       string
	Ctime      int64
	Utime      int64
}

func toProfileRow(row dao.GetMyProfileRow) ProfileRow {
	return ProfileRow{
		UserID:     row.UserID,
		Account:    row.Account,
		Email:      row.Email,
		Username:   row.Username,
		Gender:     row.Gender,
		Avatar:     row.Avatar,
		Phone:      row.Phone,
		Occupation: row.Occupation,
		Birthday:   row.Birthday,
		Province:   row.Province,
		City:       row.City,
		Ctime:      row.Ctime,
		Utime:      row.Utime,
	}
}

func (ur *UserRepo) GetMyProfile(ctx context.Context, userID string) (ProfileRow, error) {
	key := cache.ProfileKey(userID)
	{
		var cached ProfileRow
		if ok, _ := cache.GetJSON(ctx, ur.rdb, key, &cached); ok {
			ur.logCacheHit("user_cache_hit", "type", "profile", "user", userID)
			return cached, nil
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return ProfileRow{}, err
	}
	p, err := ur.userDao.GetMyProfile(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProfileRow{}, ErrUserNotExist
		}
		ur.log.Error(err)
		return ProfileRow{}, ErrDefault
	}
	ret := toProfileRow(p)
	_ = cache.SetJSON(ctx, ur.rdb, key, ret, time.Duration(userconstant.ProfileTTL)*time.Second)
	return ret, nil
}

// UpdateGender 事务地同时更新 user_base 和 user_addition 的性别，保证两表一致
func (ur *UserRepo) UpdateGender(ctx context.Context, userID, gender string) error {
	tx, err := ur.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	uq := ur.userDao.WithTx(tx)
	var uid pgtype.UUID
	if scanErr := uid.Scan(userID); scanErr != nil {
		return scanErr
	}
	n, err := uq.UpdateBaseGenderByUserID(ctx, dao.UpdateBaseGenderByUserIDParams{
		UserID: uid,
		Gender: gender,
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	n, err = uq.UpdateGenderByUserID(ctx, dao.UpdateGenderByUserIDParams{
		UserID: uid,
		Gender: gender,
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return tx.Commit(ctx)
}

func (ur *UserRepo) UpdateAvatarByID(ctx context.Context, userID, url string) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return err
	}
	// 兼容：使用 UpdateUserAdditionByUserID 仅更新 avatar 字段
	var v2, v3, v4, v5 interface{}
	v6 := url
	count, err := ur.userDao.UpdateUserAdditionByUserID(ctx, dao.UpdateUserAdditionByUserIDParams{
		UserID:  userUUID,
		Column2: v2,
		Column3: v3,
		Column4: v4,
		Column5: v5,
		Column6: v6,
		Column7: -1,
		Utime:   time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrUserUpdateFailed
	}
	if count == 0 {
		return ErrUserNotExist
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(userID))
	return nil
}

func (ur *UserRepo) UpdateAdditionByID(ctx context.Context, userID string, occupation, phone, province, city, avatar *string, birthday *int64) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return err
	}
	var v2 any
	if occupation != nil {
		v2 = *occupation
	}
	var v3 any
	if phone != nil {
		v3 = *phone
	}
	var v4 any
	if province != nil {
		v4 = *province
	}
	var v5 any
	if city != nil {
		v5 = *city
	}
	var v6 any
	if avatar != nil {
		v6 = *avatar
	}
	var v7 int64 = -1
	if birthday != nil {
		v7 = *birthday
	}
	n, err := ur.userDao.UpdateUserAdditionByUserID(ctx, dao.UpdateUserAdditionByUserIDParams{
		UserID:  userUUID,
		Column2: v2,
		Column3: v3,
		Column4: v4,
		Column5: v5,
		Column6: v6,
		Column7: v7,
		Utime:   time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(userID))
	return nil
}
