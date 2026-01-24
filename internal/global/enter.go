package global

import (
	"nurture/internal/pkg/miniox"
	"nurture/internal/pkg/pgsqlx"
	"nurture/internal/pkg/redisx"
	"nurture/internal/pkg/zapx"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

var (
	Log *zap.SugaredLogger
	DB  *pgxpool.Pool
	RDB redis.Cmdable
	MIO *minio.Client
)

func Init() {
	Log = zapx.InitZap()
	DB = pgsqlx.InitPgsql()
	RDB = redisx.InitRedis()
	MIO = miniox.InitMinio()
}
