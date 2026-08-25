package baby

import (
	"nurture/internal/baby/handler"
	"nurture/internal/baby/logic"
	"nurture/internal/baby/repo"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Deps struct {
	DB            *pgxpool.Pool
	RDB           redis.Cmdable
	Log           *zap.SugaredLogger
	PartnerReader logic.PartnerReader
}

type Module struct {
	handler *handler.BabyHandler
	client  *Client
}

func NewModule(deps Deps) *Module {
	babyRepo := repo.NewBabyRepo(deps.DB, deps.RDB, deps.Log)
	babyLogic := logic.NewBabyLogic(babyRepo, deps.PartnerReader, deps.Log)
	return &Module{
		handler: handler.NewBabyHandler(babyLogic),
		client:  NewClient(babyRepo),
	}
}

func (m *Module) Client() *Client {
	return m.client
}
