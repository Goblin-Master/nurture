package repo

import (
	"context"
	"errors"
	"nurture/internal/global"
	"nurture/internal/repo/user"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type IUserRepo interface {
	LoginWithAccount(ctx context.Context, account string, password string) (user.UserBase, error)
	LoginWithEmail(ctx context.Context, email string) (user.UserBase, error)
	Register(ctx context.Context, userID, username, email, account, password, gender string) error //注册仅写user_base，同时创建user_addition
	ResetPassword(ctx context.Context, email, newPassword string) error
	UpdateAvatarByID(ctx context.Context, userID, url string) error
}
type UserRepo struct {
	userDao *user.Queries
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		userDao: user.New(global.DB),
	}
}

var _ IUserRepo = (*UserRepo)(nil)

func (ur *UserRepo) LoginWithAccount(ctx context.Context, account string, password string) (user.UserBase, error) {
	u, err := ur.userDao.GetUserByAccountAndPassword(ctx, user.GetUserByAccountAndPasswordParams{
		Account:  account,
		Password: password,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.UserBase{}, ErrUserNotExist
		}
		global.Log.Error(err)
		return user.UserBase{}, ErrDefault
	}
	return u, nil
}

func (ur *UserRepo) LoginWithEmail(ctx context.Context, email string) (user.UserBase, error) {
	u, err := ur.userDao.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.UserBase{}, ErrUserNotExist
		}
		global.Log.Error(err)
		return user.UserBase{}, ErrDefault
	}
	return u, nil
}

func (ur *UserRepo) Register(ctx context.Context, userID, username, email, account, password, gender string) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return err
	}

	err := ur.userDao.CreateUser(ctx, user.CreateUserParams{
		UserID:   userUUID,
		Username: username,
		Email:    email,
		Account:  account,
		Password: password,
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
		global.Log.Error(err)
		return ErrDefault
	}
	return nil
}

func (ur *UserRepo) ResetPassword(ctx context.Context, email, newPassword string) error {
	count, err := ur.userDao.UpdatePasswordByEmail(ctx, user.UpdatePasswordByEmailParams{
		Email:    email,
		Password: newPassword,
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if count == 0 {
		return ErrUserNotExist
	}
	return nil
}

func (ur *UserRepo) UpdateAvatarByID(ctx context.Context, userID, url string) error {
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return err
	}
	count, err := ur.userDao.UpdateAvatarByUserID(ctx, user.UpdateAvatarByUserIDParams{
		UserID: userUUID,
		Avatar: url,
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if count == 0 {
		return ErrUserNotExist
	}
	return nil
}
