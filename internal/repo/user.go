package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nurture/internal/constant"
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
	GetMyProfile(ctx context.Context, userID string) (user.GetMyProfileRow, error)
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
	BindPartnerAndSyncBabies(ctx context.Context, fatherUserID, motherUserID string) error
	UpdateGender(ctx context.Context, userID, gender string) error
	FollowUser(ctx context.Context, followerID, followeeID string) error
	UnfollowUser(ctx context.Context, followerID, followeeID string) error
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]user.ListFollowingByUserIDRow, bool, error)
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]user.ListFollowersByUserIDRow, bool, error)
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

func (ur *UserRepo) GetMyProfile(ctx context.Context, userID string) (user.GetMyProfileRow, error) {
	if global.RDB != nil {
		key := fmt.Sprintf(constant.USER_PROFILE_KEY, userID)
		if s, err := global.RDB.Get(ctx, key).Result(); err == nil && s != "" {
			var cached user.GetMyProfileRow
			if jsonErr := json.Unmarshal([]byte(s), &cached); jsonErr == nil {
				global.Log.Info("user_cache_hit", "type", "profile", "user", userID)
				return cached, nil
			}
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return user.GetMyProfileRow{}, err
	}
	p, err := ur.userDao.GetMyProfile(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.GetMyProfileRow{}, ErrUserNotExist
		}
		global.Log.Error(err)
		return user.GetMyProfileRow{}, ErrDefault
	}
	if global.RDB != nil {
		if b, mErr := json.Marshal(p); mErr == nil {
			_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, userID), string(b), time.Duration(constant.USER_PROFILE_TTL)*time.Second).Err()
		}
	}
	return p, nil
}

func (ur *UserRepo) FollowUser(ctx context.Context, followerID, followeeID string) error {
	var fUID, eUID pgtype.UUID
	if err := fUID.Scan(followerID); err != nil {
		return err
	}
	if err := eUID.Scan(followeeID); err != nil {
		return err
	}
	err := ur.userDao.CreateFollow(ctx, user.CreateFollowParams{
		Follower: fUID,
		Followee: eUID,
		Ctime:    time.Now().UnixMilli(),
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, followerID)).Err()
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, followeeID)).Err()
		iter1 := global.RDB.Scan(ctx, 0, fmt.Sprintf(constant.USER_FOLLOWING_KEY, followerID, 0, 0), 100).Iterator()
		for iter1.Next(ctx) {
			_ = global.RDB.Del(ctx, iter1.Val()).Err()
		}
		iter2 := global.RDB.Scan(ctx, 0, fmt.Sprintf(constant.USER_FOLLOWERS_KEY, followeeID, 0, 0), 100).Iterator()
		for iter2.Next(ctx) {
			_ = global.RDB.Del(ctx, iter2.Val()).Err()
		}
	}
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
	n, err := ur.userDao.DeleteFollow(ctx, user.DeleteFollowParams{
		Follower: fUID,
		Followee: eUID,
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if n == 0 {
		return nil
	}
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, followerID)).Err()
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, followeeID)).Err()
		iter1 := global.RDB.Scan(ctx, 0, fmt.Sprintf("user:following:%s:*", followerID), 100).Iterator()
		for iter1.Next(ctx) {
			_ = global.RDB.Del(ctx, iter1.Val()).Err()
		}
		iter2 := global.RDB.Scan(ctx, 0, fmt.Sprintf("user:followers:%s:*", followeeID), 100).Iterator()
		for iter2.Next(ctx) {
			_ = global.RDB.Del(ctx, iter2.Val()).Err()
		}
	}
	return nil
}

func (ur *UserRepo) ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]user.ListFollowingByUserIDRow, bool, error) {
	if global.RDB != nil {
		key := fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, pageSize)
		if s, err := global.RDB.Get(ctx, key).Result(); err == nil && s != "" {
			type listCache struct {
				Rows    []user.ListFollowingByUserIDRow `json:"rows"`
				HasMore bool                            `json:"has_more"`
			}
			var cached listCache
			if jsonErr := json.Unmarshal([]byte(s), &cached); jsonErr == nil {
				global.Log.Info("user_cache_hit", "type", "following", "user", userID, "page", page, "size", pageSize)
				return cached.Rows, cached.HasMore, nil
			}
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, err
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.ListFollowingByUserID(ctx, user.ListFollowingByUserIDParams{
		Follower: uid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, pageSize)).Err()
	}
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	if global.RDB != nil {
		type listCache struct {
			Rows    []user.ListFollowingByUserIDRow `json:"rows"`
			HasMore bool                            `json:"has_more"`
		}
		payload := listCache{Rows: rows, HasMore: hasMore}
		if b, mErr := json.Marshal(payload); mErr == nil {
			_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_FOLLOWING_KEY, userID, page, pageSize), string(b), time.Duration(constant.USER_LIST_TTL)*time.Second).Err()
		}
	}
	return rows, hasMore, nil
}

func (ur *UserRepo) ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]user.ListFollowersByUserIDRow, bool, error) {
	if global.RDB != nil {
		key := fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, pageSize)
		if s, err := global.RDB.Get(ctx, key).Result(); err == nil && s != "" {
			type listCache struct {
				Rows    []user.ListFollowersByUserIDRow `json:"rows"`
				HasMore bool                            `json:"has_more"`
			}
			var cached listCache
			if jsonErr := json.Unmarshal([]byte(s), &cached); jsonErr == nil {
				global.Log.Info("user_cache_hit", "type", "followers", "user", userID, "page", page, "size", pageSize)
				return cached.Rows, cached.HasMore, nil
			}
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, false, err
	}
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.ListFollowersByUserID(ctx, user.ListFollowersByUserIDParams{
		Followee: uid,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, pageSize)).Err()
	}
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	if global.RDB != nil {
		type listCache struct {
			Rows    []user.ListFollowersByUserIDRow `json:"rows"`
			HasMore bool                            `json:"has_more"`
		}
		payload := listCache{Rows: rows, HasMore: hasMore}
		if b, mErr := json.Marshal(payload); mErr == nil {
			_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_FOLLOWERS_KEY, userID, page, pageSize), string(b), time.Duration(constant.USER_LIST_TTL)*time.Second).Err()
		}
	}
	return rows, hasMore, nil
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
	if global.RDB != nil {
		key := fmt.Sprintf(constant.USER_PARTNER_KEY, userID)
		if s, err := global.RDB.Get(ctx, key).Result(); err == nil {
			global.Log.Info("user_cache_hit", "type", "partner", "user", userID)
			return s, nil
		}
	}
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := ur.userDao.GetPartnerByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if global.RDB != nil {
				_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_PARTNER_KEY, userID), "", time.Duration(constant.USER_PROFILE_TTL)*time.Second).Err()
			}
			return "", nil
		}
		global.Log.Error(err)
		return "", ErrDefault
	}
	if row.Father == userID {
		if global.RDB != nil {
			_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_PARTNER_KEY, userID), row.Mother, time.Duration(constant.USER_PARTNER_TTL)*time.Second).Err()
		}
		return row.Mother, nil
	}
	if global.RDB != nil {
		_ = global.RDB.SetEX(ctx, fmt.Sprintf(constant.USER_PARTNER_KEY, userID), row.Father, time.Duration(constant.USER_PARTNER_TTL)*time.Second).Err()
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
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PARTNER_KEY, fatherUserID)).Err()
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PARTNER_KEY, motherUserID)).Err()
		for _, uid := range []string{fatherUserID, motherUserID} {
			iter1 := global.RDB.Scan(ctx, 0, fmt.Sprintf("user:following:%s:*", uid), 100).Iterator()
			for iter1.Next(ctx) {
				_ = global.RDB.Del(ctx, iter1.Val()).Err()
			}
			iter2 := global.RDB.Scan(ctx, 0, fmt.Sprintf("user:followers:%s:*", uid), 100).Iterator()
			for iter2.Next(ctx) {
				_ = global.RDB.Del(ctx, iter2.Val()).Err()
			}
			_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, uid)).Err()
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
	// 兼容：使用 UpdateUserAdditionByUserID 仅更新 avatar 字段
	var v2, v3, v4, v5 interface{}
	v6 := url
	count, err := ur.userDao.UpdateUserAdditionByUserID(ctx, user.UpdateUserAdditionByUserIDParams{
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
		global.Log.Error(err)
		return ErrUserUpdateFailed
	}
	if count == 0 {
		return ErrUserNotExist
	}
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, userID)).Err()
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
	if global.RDB != nil {
		_ = global.RDB.Del(ctx, fmt.Sprintf(constant.USER_PROFILE_KEY, userID)).Err()
	}
	return nil
}

// admin
func (ur *UserRepo) AdminListUsers(ctx context.Context, keyword string, page, pageSize int) ([]user.AdminListUsersRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.AdminListUsers(ctx, user.AdminListUsersParams{
		Column1: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	return rows, hasMore, nil
}

func (ur *UserRepo) AdminUpdateUserRole(ctx context.Context, userID string, role int16) error {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}
	n, err := ur.userDao.AdminUpdateUserRole(ctx, user.AdminUpdateUserRoleParams{
		UserID: uid,
		Role:   role,
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		global.Log.Error(err)
		return ErrDefault
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return nil
}
