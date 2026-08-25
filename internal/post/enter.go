package post

import (
	"nurture/internal/pkg/aix"
	"nurture/internal/post/handler"
	"nurture/internal/post/logic"
	"nurture/internal/post/repo"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Deps struct {
	DB           *pgxpool.Pool
	RDB          redis.Cmdable
	Log          *zap.SugaredLogger
	AI           *aix.AIX
	FollowReader logic.FollowReader
}

type Module struct {
	handler *handler.PostHandler
}

func NewModule(deps Deps) *Module {
	postRepo := repo.NewPostRepo(deps.DB, deps.RDB, deps.Log, deps.AI)
	postLogic := logic.NewPostLogic(postRepo, deps.FollowReader, deps.Log)
	return &Module{
		handler: handler.NewPostHandler(postLogic),
	}
}
