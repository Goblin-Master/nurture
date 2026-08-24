package chat

import (
	"context"
	"nurture/internal/chat/handler"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/repo"
	"nurture/internal/chat/session"
	"nurture/internal/chat/worker"
	"nurture/internal/pkg/rabbitmqx"
	"nurture/internal/pkg/ratelimitx"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type RateLimitUserFunc func(key string, limit int64, window time.Duration) gin.HandlerFunc

type Deps struct {
	DB            *pgxpool.Pool
	RDB           redis.Cmdable
	RabbitMQ      *rabbitmqx.Client
	Log           *zap.SugaredLogger
	AuthUser      gin.HandlerFunc
	RateLimitUser RateLimitUserFunc
}

type Module struct {
	handler       *handler.ChatHandler
	authUser      gin.HandlerFunc
	rateLimitUser RateLimitUserFunc
}

func NewModule(deps Deps) *Module {
	authUser := deps.AuthUser
	if authUser == nil {
		authUser = noopMiddleware
	}
	rateLimitUser := deps.RateLimitUser
	if rateLimitUser == nil {
		rateLimitUser = noopRateLimit
	}
	hub := session.NewHub()
	go hub.Run()
	chatRepo := repo.NewChatRepo(deps.DB, deps.RDB, deps.Log)
	if deps.RabbitMQ != nil {
		worker.NewWorker(deps.RabbitMQ, hub, deps.Log).Start(context.Background())
	}
	if deps.DB != nil && deps.RabbitMQ != nil {
		worker.NewOutboxWorker(chatRepo, deps.RabbitMQ, deps.Log).Start(context.Background())
	}
	chatLogic := logic.NewChatLogic(chatRepo, ratelimitx.NewLimiter(deps.RDB))

	return &Module{
		handler:       handler.NewChatHandler(chatLogic, hub),
		authUser:      authUser,
		rateLimitUser: rateLimitUser,
	}
}

func noopMiddleware(c *gin.Context) {
	c.Next()
}

func noopRateLimit(string, int64, time.Duration) gin.HandlerFunc {
	return noopMiddleware
}
