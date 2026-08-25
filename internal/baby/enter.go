package baby

import (
	"context"
	"nurture/internal/baby/handler"
	"nurture/internal/baby/logic"
	"nurture/internal/baby/repo"
	"nurture/internal/baby/worker"
	"nurture/internal/pkg/rabbitmqx"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Deps struct {
	DB            *pgxpool.Pool
	RDB           redis.Cmdable
	RabbitMQ      *rabbitmqx.Client
	Log           *zap.SugaredLogger
	PartnerReader logic.PartnerReader
	Context       context.Context
}

type Module struct {
	handler *handler.BabyHandler
	client  *Client
}

func NewModule(deps Deps) *Module {
	ctx := deps.Context
	if ctx == nil {
		ctx = context.Background()
	}
	babyRepo := repo.NewBabyRepo(deps.DB, deps.RDB, deps.Log)
	babyLogic := logic.NewBabyLogic(babyRepo, deps.PartnerReader, deps.Log)
	if deps.DB != nil && deps.RabbitMQ != nil {
		eventLogic := logic.NewBabyEventLogic(babyRepo, deps.Log)
		worker.NewWorker(deps.RabbitMQ, eventLogic, deps.Log).Start(ctx)
	}
	return &Module{
		handler: handler.NewBabyHandler(babyLogic),
		client:  NewClient(babyRepo),
	}
}

func (m *Module) Client() *Client {
	return m.client
}
