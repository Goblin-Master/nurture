package chat

import (
	"errors"
	"nurture/internal/chat/handler"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/repo"
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
	Log           *zap.SugaredLogger
	AuthUser      gin.HandlerFunc
	RateLimitUser RateLimitUserFunc
	GetUserID     handler.GetUserIDFunc
	ParseToken    handler.ParseTokenFunc
	Respond       handler.RespondFunc
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
	getUserID := deps.GetUserID
	if getUserID == nil {
		getUserID = defaultGetUserID
	}
	parseToken := deps.ParseToken
	if parseToken == nil {
		parseToken = defaultParseToken
	}
	respond := deps.Respond
	if respond == nil {
		respond = defaultRespond
	}

	chatRepo := repo.NewChatRepo(deps.DB, deps.RDB, deps.Log)
	chatLogic := logic.NewChatLogic(chatRepo, ratelimitx.NewLimiter(deps.RDB))

	return &Module{
		handler:       handler.NewChatHandler(chatLogic, getUserID, parseToken, respond),
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

func defaultGetUserID(*gin.Context) string {
	return ""
}

func defaultParseToken(string) (string, error) {
	return "", errors.New("token无效")
}

func defaultRespond(c *gin.Context, resp interface{}, err error) {
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "OK", "data": resp})
}
