package repo

import (
	"context"
	"errors"
	"nurture/internal/user/repo/cache"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (ur *UserRepo) IsPhoneUsed(ctx context.Context, phone string, excludeUserID string) (bool, error) {
	var exclude pgtype.UUID
	if err := exclude.Scan(excludeUserID); err != nil {
		return false, err
	}
	exists, err := ur.userDao.IsPhoneUsed(ctx, dao.IsPhoneUsedParams{
		Phone:  phone,
		UserID: exclude,
	})
	if err != nil {
		ur.log.Error(err)
		return false, ErrDefault
	}
	return exists, nil
}

func (ur *UserRepo) BindEmail(ctx context.Context, userID, email string) error {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}
	count, err := ur.userDao.BindEmailByUserID(ctx, dao.BindEmailByUserIDParams{
		UserID: uid,
		Email:  pgtype.Text{String: email, Valid: true},
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "user_base_email_key":
				return ErrEmailIsUsed
			}
		}
		ur.log.Error(err)
		return ErrUserUpdateFailed
	}
	if count == 0 {
		return ErrUserNotExist
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(userID))
	return nil
}
