package ai

import (
	"nurture/internal/ai/handler"
	"nurture/internal/ai/logic"
	"nurture/internal/ai/repo"
	"nurture/internal/config"
	"nurture/internal/pkg/aix"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Deps struct {
	RDB          redis.Cmdable
	Log          *zap.SugaredLogger
	AI           *aix.AIX
	AIConfig     config.AI
	DBEnabled    bool
	GrowthReader logic.BabyGrowthReader
}

type Module struct {
	handler *handler.AIHandler
}

func NewModule(deps Deps) *Module {
	aiRepo := repo.NewAIRepo(deps.AI, deps.RDB, deps.Log)
	aiLogic := logic.NewAILogic(aiRepo, deps.AI, deps.AIConfig, deps.GrowthReader, deps.DBEnabled, deps.Log)
	return &Module{
		handler: handler.NewAIHandler(aiLogic, deps.Log),
	}
}
