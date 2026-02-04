package global

import (
	"nurture/internal/config"
	"nurture/internal/pkg/aix"
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
	AIX *aix.AIX // AI 功能实例
)

func Init() {
	Log = zapx.InitZap()
	DB = pgsqlx.InitPgsql()
	RDB = redisx.InitRedis()
	MIO = miniox.InitMinio()

	// 初始化 AIX
	var err error
	AIX, err = aix.NewAIX(config.Conf.AI, RDB, config.Conf.DB.DSN())
	if err != nil {
		panic("AIX init failed: " + err.Error())
	}
}
