package repo

import (
	"context"
	"errors"
	"nurture/internal/global"
	"nurture/internal/repo/baby"
	"nurture/internal/repo/user"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type IUserRepo interface {
	LoginWithAccount(ctx context.Context, account string, password string) (user.UserBase, error)
	LoginWithEmail(ctx context.Context, email string) (user.UserBase, error)
	GetUserByID(ctx context.Context, userID string) (user.UserBase, error)
	Register(ctx context.Context, userID, username, email, account, password, gender string) error //注册仅写user_base，同时创建user_addition
	ResetPassword(ctx context.Context, email, newPassword string) error
	UpdateAvatarByID(ctx context.Context, userID, url string) error
	UpdateAdditionByID(ctx context.Context, userID string, occupation, phone, province, city, avatar *string, birthday *int64) error
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
	BindPartnerAndSyncBabies(ctx context.Context, fatherUserID, motherUserID string) error
	UpdateGender(ctx context.Context, userID, gender string) error
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

func (ur *UserRepo) GetUserByID(ctx context.Context, userID string) (user.UserBase, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return user.UserBase{}, err
	}
	u, err := ur.userDao.GetUserByID(ctx, uid)
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

// UpdateGender 事务地同时更新 user_base 和 user_addition 的性别，保证两表一致
func (ur *UserRepo) UpdateGender(ctx context.Context, userID, gender string) error {
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()
	uq := ur.userDao.WithTx(tx)
	var uid pgtype.UUID
	if scanErr := uid.Scan(userID); scanErr != nil {
		err = scanErr
		return err
	}
	if _, err = uq.UpdateBaseGenderByUserID(ctx, user.UpdateBaseGenderByUserIDParams{
		UserID: uid,
		Gender: gender,
		Utime:  time.Now().UnixMilli(),
	}); err != nil {
		global.Log.Error(err)
		return ErrUserUpdateFailed
	}
	if _, err = uq.UpdateGenderByUserID(ctx, user.UpdateGenderByUserIDParams{
		UserID: uid,
		Gender: gender,
		Utime:  time.Now().UnixMilli(),
	}); err != nil {
		global.Log.Error(err)
		return ErrUserUpdateFailed
	}
	return nil
}

// GetPartnerByUserID 查询另一半（返回对方ID；没有则返回空字符串）
func (ur *UserRepo) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := ur.userDao.GetPartnerByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		global.Log.Error(err)
		return "", ErrDefault
	}
	if row.Father == userID {
		return row.Mother, nil
	}
	return row.Father, nil
}

// BindPartnerAndSyncBabies 建立关系并双向同步现有宝宝
func (ur *UserRepo) BindPartnerAndSyncBabies(ctx context.Context, fatherUserID, motherUserID string) error {
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	// 1) 插入关系（幂等）使用 sqlc
	uq := ur.userDao.WithTx(tx)
	var aUUID, bUUID pgtype.UUID
	if scanErr := aUUID.Scan(fatherUserID); scanErr != nil {
		return scanErr
	}
	if scanErr := bUUID.Scan(motherUserID); scanErr != nil {
		return scanErr
	}
	if err = uq.CreatePartner(ctx, user.CreatePartnerParams{
		Father: aUUID,
		Mother: bUUID,
		Ctime:  time.Now().UnixMilli(),
	}); err != nil {
		global.Log.Error(err)
		return ErrDefault
	}

	// 2) 同步宝宝：A -> B
	bq := baby.New(tx)
	aBabies, err := bq.ListBabiesByUserID(ctx, aUUID)
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	for _, row := range aBabies {
		// 获取完整信息
		babyID := pgtype.UUID{}
		if err := babyID.Scan(row.BabyID); err != nil {
			return err
		}
		full, err := bq.GetBabyByIDAndUser(ctx, baby.GetBabyByIDAndUserParams{
			BabyID: babyID,
			UserID: aUUID,
		})
		if err != nil {
			global.Log.Error(err)
			return ErrDefault
		}
		// 插入到 B（幂等：若已存在则忽略）
		err = bq.CreateBaby(ctx, baby.CreateBabyParams{
			BabyID:   full.BabyID,
			UserID:   bUUID,
			Name:     full.Name,
			Gender:   full.Gender,
			Birthday: full.Birthday,
			Avatar:   full.Avatar,
			Ctime:    time.Now().UnixMilli(),
			Utime:    time.Now().UnixMilli(),
		})
		if err != nil {
			global.Log.Error(err)
			return ErrDefault
		}
	}

	// 3) 同步宝宝：B -> A
	bBabies, err := bq.ListBabiesByUserID(ctx, bUUID)
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	for _, row := range bBabies {
		babyID := pgtype.UUID{}
		if err := babyID.Scan(row.BabyID); err != nil {
			return err
		}
		full, err := bq.GetBabyByIDAndUser(ctx, baby.GetBabyByIDAndUserParams{
			BabyID: babyID,
			UserID: bUUID,
		})
		if err != nil {
			global.Log.Error(err)
			return ErrDefault
		}
		err = bq.CreateBaby(ctx, baby.CreateBabyParams{
			BabyID:   full.BabyID,
			UserID:   aUUID,
			Name:     full.Name,
			Gender:   full.Gender,
			Birthday: full.Birthday,
			Avatar:   full.Avatar,
			Ctime:    time.Now().UnixMilli(),
			Utime:    time.Now().UnixMilli(),
		})
		if err != nil {
			global.Log.Error(err)
			return ErrDefault
		}
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
		return ErrUserUpdateFailed
	}
	if count == 0 {
		return ErrUserNotExist
	}
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
	n, err := ur.userDao.UpdateUserAdditionByUserID(ctx, user.UpdateUserAdditionByUserIDParams{
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
		global.Log.Error(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return nil
}
