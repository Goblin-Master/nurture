package repo

import (
	"context"
	"errors"
	"nurture/internal/pkg/passwordx"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserBaseRow struct {
	UserID   string
	Account  string
	Password string
	Email    string
	Username string
	Gender   string
	Role     int16
	Ctime    int64
	Utime    int64
}

func toUserBaseRow(row dao.UserBase) UserBaseRow {
	return UserBaseRow{
		UserID:   row.UserID.String(),
		Account:  row.Account,
		Password: row.Password,
		Email:    row.Email.String,
		Username: row.Username,
		Gender:   row.Gender,
		Role:     row.Role,
		Ctime:    row.Ctime,
		Utime:    row.Utime,
	}
}

func (ur *UserRepo) LoginWithAccount(ctx context.Context, account string, password string) (UserBaseRow, error) {
	u, err := ur.userDao.GetUserByAccount(ctx, account)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserBaseRow{}, ErrUserNotExist
		}
		ur.log.Error(err)
		return UserBaseRow{}, ErrDefault
	}
	if passwordx.IsBcryptHash(u.Password) {
		if err := passwordx.ComparePassword(u.Password, password); err != nil {
			if errors.Is(err, passwordx.ErrPasswordMismatch) || errors.Is(err, passwordx.ErrPasswordEmpty) {
				return UserBaseRow{}, ErrAccountOrPwd
			}
			ur.log.Error(err)
			return UserBaseRow{}, ErrDefault
		}
		return toUserBaseRow(u), nil
	}
	if password == "" || u.Password == "" || u.Password != password {
		return UserBaseRow{}, ErrAccountOrPwd
	}
	if hashed, err := passwordx.HashAnyPassword(password); err == nil {
		if _, upErr := ur.userDao.UpdatePasswordByUserID(ctx, dao.UpdatePasswordByUserIDParams{
			UserID:   u.UserID,
			Password: hashed,
			Utime:    time.Now().UnixMilli(),
		}); upErr != nil {
			ur.log.Error(upErr)
		}
	} else {
		ur.log.Error(err)
	}
	return toUserBaseRow(u), nil
}

func (ur *UserRepo) LoginWithEmail(ctx context.Context, email string) (UserBaseRow, error) {
	u, err := ur.userDao.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserBaseRow{}, ErrUserNotExist
		}
		ur.log.Error(err)
		return UserBaseRow{}, ErrDefault
	}
	return toUserBaseRow(u), nil
}

func (ur *UserRepo) GetUserByID(ctx context.Context, userID string) (UserBaseRow, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return UserBaseRow{}, err
	}
	u, err := ur.userDao.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserBaseRow{}, ErrUserNotExist
		}
		ur.log.Error(err)
		return UserBaseRow{}, ErrDefault
	}
	return toUserBaseRow(u), nil
}

func (ur *UserRepo) Register(ctx context.Context, userID, username string, email *string, account, password, gender string) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return err
	}
	hashed, err := passwordx.HashPassword(password)
	if err != nil {
		return err
	}
	var emailText pgtype.Text
	if email != nil && *email != "" {
		emailText = pgtype.Text{String: *email, Valid: true}
	}

	err = ur.userDao.CreateUser(ctx, dao.CreateUserParams{
		UserID:   userUUID,
		Username: username,
		Email:    emailText,
		Account:  account,
		Password: hashed,
		Ctime:    time.Now().UnixMilli(),
		Utime:    time.Now().UnixMilli(),
		Gender:   gender,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			switch pgErr.ConstraintName {
			case "user_base_account_key":
				return ErrAccountIsUsed
			case "user_base_email_key":
				return ErrEmailIsUsed
			}
		}
		ur.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (ur *UserRepo) ResetPassword(ctx context.Context, email, newPassword string) error {
	hashed, err := passwordx.HashPassword(newPassword)
	if err != nil {
		return err
	}
	count, err := ur.userDao.UpdatePasswordByEmail(ctx, dao.UpdatePasswordByEmailParams{
		Email:    pgtype.Text{String: email, Valid: true},
		Password: hashed,
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	if count == 0 {
		return ErrUserNotExist
	}
	return nil
}
