package user

import (
	"context"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/smsx"
	"nurture/internal/user/handler"
	"nurture/internal/user/logic"
	"nurture/internal/user/repo"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Deps struct {
	DB         *pgxpool.Pool
	RDB        redis.Cmdable
	Log        *zap.SugaredLogger
	Email      logic.EmailSender
	SMS        logic.SMSSender
	BabySyncer logic.BabySyncer
}

type Module struct {
	userRepo repo.IUserRepo
	handler  *handler.UserHandler
}

func NewModule(deps Deps) *Module {
	email := deps.Email
	if email == nil {
		email = emailx.NewEmailX()
	}
	sms := deps.SMS
	if sms == nil {
		sms = smsx.NewSmsX()
	}
	userRepo := repo.NewUserRepo(deps.DB, deps.RDB, deps.Log)
	userLogic := logic.NewUserLogic(userRepo, email, sms, deps.BabySyncer, deps.Log)
	return &Module{
		userRepo: userRepo,
		handler:  handler.NewUserHandler(userLogic),
	}
}

func (m *Module) GetPartnerByUserID(ctx context.Context, userID string) (string, error) {
	return m.userRepo.GetPartnerByUserID(ctx, userID)
}

func (m *Module) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	return m.userRepo.IsFollowing(ctx, followerID, followeeID)
}
