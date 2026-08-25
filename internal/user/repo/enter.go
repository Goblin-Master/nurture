package repo

import (
	"context"
	"nurture/internal/pkg/zapx"
	"nurture/internal/user/repo/dao"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

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
