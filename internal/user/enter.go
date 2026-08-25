package user

import (
	"context"
	"nurture/internal/pkg/emailx"
	"nurture/internal/pkg/rabbitmqx"
	"nurture/internal/pkg/smsx"
	"nurture/internal/user/handler"
	"nurture/internal/user/logic"
	"nurture/internal/user/repo"
	"nurture/internal/user/worker"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Deps struct {
	DB       *pgxpool.Pool
	RDB      redis.Cmdable
	RabbitMQ *rabbitmqx.Client
	Log      *zap.SugaredLogger
	Email    emailx.Sender
	SMS      smsx.Sender
	Context  context.Context
}

type Module struct {
	userLogic *logic.UserLogic
	handler   *handler.UserHandler
	client    *Client
}

func NewModule(deps Deps) *Module {
	ctx := deps.Context
	if ctx == nil {
		ctx = context.Background()
	}
	email := deps.Email
	if email == nil {
		email = emailx.NewEmailX()
	}
	sms := deps.SMS
	if sms == nil {
		sms = smsx.NewSmsX()
	}
	userRepo := repo.NewUserRepo(deps.DB, deps.RDB, deps.Log)
	if deps.DB != nil && deps.RabbitMQ != nil {
		worker.NewOutboxWorker(userRepo, deps.RabbitMQ, deps.Log).Start(ctx)
	}
	userLogic := logic.NewUserLogic(userRepo, email, sms, deps.Log)
	return &Module{
		userLogic: userLogic,
		handler:   handler.NewUserHandler(userLogic),
		client:    NewClient(userRepo),
	}
}

func (m *Module) Client() *Client {
	return m.client
}
