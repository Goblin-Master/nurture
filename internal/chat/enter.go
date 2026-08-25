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
	Context       context.Context
}

type Module struct {
	handler       *handler.ChatHandler
	authUser      gin.HandlerFunc
	rateLimitUser RateLimitUserFunc
	cancel        context.CancelFunc
	done          <-chan struct{}
}

func NewModule(deps Deps) *Module {
	ctx := deps.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	authUser := deps.AuthUser
	if authUser == nil {
		authUser = noopMiddleware
	}
	rateLimitUser := deps.RateLimitUser
	if rateLimitUser == nil {
		rateLimitUser = noopRateLimit
	}
	hub := session.NewHub()
	go hub.Run(ctx)
	chatRepo := repo.NewChatRepo(deps.DB, deps.RDB, deps.Log)
	if deps.RabbitMQ != nil {
		worker.NewWorker(deps.RabbitMQ, hub, deps.Log).Start(ctx)
	}
	if deps.DB != nil && deps.RabbitMQ != nil {
		worker.NewOutboxWorker(chatRepo, deps.RabbitMQ, deps.Log).Start(ctx)
	}
	chatLogic := logic.NewChatLogic(chatRepo, ratelimitx.NewLimiter(deps.RDB))

	return &Module{
		handler:       handler.NewChatHandler(chatLogic, hub),
		authUser:      authUser,
		rateLimitUser: rateLimitUser,
		cancel:        cancel,
		done:          hub.Done(),
	}
}

func (m *Module) Close() {
	if m == nil || m.cancel == nil {
		return
	}
	m.cancel()
}

func (m *Module) Done() <-chan struct{} {
	if m == nil {
		return nil
	}
	return m.done
}

func noopMiddleware(c *gin.Context) {
	c.Next()
}

func noopRateLimit(string, int64, time.Duration) gin.HandlerFunc {
	return noopMiddleware
}
