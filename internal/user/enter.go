package user

import (
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
	Email      emailx.Sender
	SMS        smsx.Sender
	BabySyncer logic.BabySyncer
}

type Module struct {
	userLogic *logic.UserLogic
	handler   *handler.UserHandler
	client    *Client
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
		userLogic: userLogic,
		handler:   handler.NewUserHandler(userLogic),
		client:    NewClient(userRepo),
	}
}

func (m *Module) Client() *Client {
	return m.client
}

func (m *Module) SetBabySyncer(syncer logic.BabySyncer) {
	m.userLogic.SetBabySyncer(syncer)
}
