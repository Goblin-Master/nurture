package global

import (
	"nurture/internal/pkg/pgsqlx"
	"nurture/internal/pkg/redisx"
	"nurture/internal/pkg/zapx"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	Log *zap.SugaredLogger
	DB  *pgxpool.Pool
	RDB redis.Cmdable
)

func Init() {
	Log = zapx.InitZap()
	DB = pgsqlx.InitPgsql()
	RDB = redisx.InitRedis()
}
