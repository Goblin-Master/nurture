package repo

import (
	"context"
	"errors"
	"nurture/internal/pkg/passwordx"
	"nurture/internal/pkg/zapx"
	userconstant "nurture/internal/user/constant"
	"nurture/internal/user/repo/cache"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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

type FollowUserRow struct {
	UserID     string
	Username   string
	Avatar     string
	FollowTime int64
}

type AdminUserRow struct {
	UserID   string
	Username string
	Avatar   string
}

type IUserRepo interface {
	LoginWithAccount(ctx context.Context, account string, password string) (UserBaseRow, error)
	LoginWithEmail(ctx context.Context, email string) (UserBaseRow, error)
	GetUserByID(ctx context.Context, userID string) (UserBaseRow, error)
	Register(ctx context.Context, userID, username string, email *string, account, password, gender string) error //注册仅写user_base，同时创建user_addition
	ResetPassword(ctx context.Context, email, newPassword string) error
	UpdateAvatarByID(ctx context.Context, userID, url string) error
	UpdateAdditionByID(ctx context.Context, userID string, occupation, phone, province, city, avatar *string, birthday *int64) error
	GetMyProfile(ctx context.Context, userID string) (ProfileRow, error)
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
	BindPartner(ctx context.Context, fatherUserID, motherUserID string) error
	UpdateGender(ctx context.Context, userID, gender string) error
	FollowUser(ctx context.Context, followerID, followeeID string) error
	UnfollowUser(ctx context.Context, followerID, followeeID string) error
	ListFollowing(ctx context.Context, userID string, page, pageSize int) ([]FollowUserRow, bool, error)
	ListFollowers(ctx context.Context, userID string, page, pageSize int) ([]FollowUserRow, bool, error)
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	IsPhoneUsed(ctx context.Context, phone string, excludeUserID string) (bool, error)
	BindEmail(ctx context.Context, userID, email string) error
	AdminListUsers(ctx context.Context, keyword string, page, pageSize int) ([]AdminUserRow, bool, error)
	AdminUpdateUserRole(ctx context.Context, userID string, role int16) error
}
type UserRepo struct {
	db      *pgxpool.Pool
	userDao *dao.Queries
	rdb     redis.Cmdable
	log     *zap.SugaredLogger
}

func NewUserRepo(db *pgxpool.Pool, rdb redis.Cmdable, log *zap.SugaredLogger) *UserRepo {
	return &UserRepo{
		db:      db,
		userDao: dao.New(db),
		rdb:     rdb,
		log:     zapx.OrNop(log),
	}
}

var _ IUserRepo = (*UserRepo)(nil)

func (ur *UserRepo) logError(err error) {
	if err != nil {
		ur.log.Error(err)
	}
}

func (ur *UserRepo) logCacheHit(args ...interface{}) {
	ur.log.Info(args...)
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

func toFollowUserRow(userID, username, avatar string, followTime int64) FollowUserRow {
	return FollowUserRow{
		UserID:     userID,
		Username:   username,
		Avatar:     avatar,
		FollowTime: followTime,
	}
}

func toAdminUserRow(row dao.AdminListUsersRow) AdminUserRow {
	return AdminUserRow{
		UserID:   row.UserID,
		Username: row.Username,
		Avatar:   row.Avatar,
	}
}

func (ur *UserRepo) LoginWithAccount(ctx context.Context, account string, password string) (UserBaseRow, error) {
	u, err := ur.userDao.GetUserByAccount(ctx, account)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserBaseRow{}, ErrUserNotExist
		}
		ur.logError(err)
		return UserBaseRow{}, ErrDefault
	}
	if passwordx.IsBcryptHash(u.Password) {
		if err := passwordx.ComparePassword(u.Password, password); err != nil {
			if errors.Is(err, passwordx.ErrPasswordMismatch) || errors.Is(err, passwordx.ErrPasswordEmpty) {
				return UserBaseRow{}, ErrAccountOrPwd
			}
			ur.logError(err)
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
			ur.logError(upErr)
		}
	} else {
		ur.logError(err)
	}
	return toUserBaseRow(u), nil
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
		ur.logError(err)
		return ProfileRow{}, ErrDefault
	}
	ret := toProfileRow(p)
	_ = cache.SetJSON(ctx, ur.rdb, key, ret, time.Duration(userconstant.ProfileTTL)*time.Second)
	return ret, nil
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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
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

func (ur *UserRepo) LoginWithEmail(ctx context.Context, email string) (UserBaseRow, error) {
	u, err := ur.userDao.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserBaseRow{}, ErrUserNotExist
		}
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
		return ErrDefault
	}
	return nil
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
		ur.logError(err)
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
		ur.logError(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return tx.Commit(ctx)
}

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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
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
		ur.logError(err)
		return ErrUserUpdateFailed
	}
	if n == 0 {
		return ErrUserNotExist
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(userID))
	return nil
}

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
		ur.logError(err)
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
		ur.logError(err)
		return ErrUserUpdateFailed
	}
	if count == 0 {
		return ErrUserNotExist
	}
	_ = cache.Del(ctx, ur.rdb, cache.ProfileKey(userID))
	return nil
}

func (ur *UserRepo) AdminListUsers(ctx context.Context, keyword string, page, pageSize int) ([]AdminUserRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.AdminListUsers(ctx, dao.AdminListUsersParams{
		Column1: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		ur.logError(err)
		return nil, false, ErrDefault
	}
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	ret := make([]AdminUserRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, toAdminUserRow(row))
	}
	return ret, hasMore, nil
}

func (ur *UserRepo) AdminUpdateUserRole(ctx context.Context, userID string, role int16) error {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}
	n, err := ur.userDao.AdminUpdateUserRole(ctx, dao.AdminUpdateUserRoleParams{
		UserID: uid,
		Role:   role,
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		ur.logError(err)
		return ErrDefault
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return nil
}
