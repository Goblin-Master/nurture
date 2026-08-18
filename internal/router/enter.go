package router

import (
	"fmt"
	"nurture/internal/chat"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// RunServer 启动服务器 路由层
func RunServer() {
	r, err := listen()
	if err != nil {
		panic(err.Error())
	}
	err = r.Run(fmt.Sprintf("%s:%d", config.Conf.App.Host, config.Conf.App.Port))
	if err != nil {
		panic(err.Error())
	}
}

// listen 配置 Gin 服务器
func listen() (*gin.Engine, error) {
	r := gin.Default()
	requestGlobalMiddleware(r)
	registerRoutes(r)
	return r, nil
}

// requestGlobalMiddleware 注册全局中间件，应用于所有路由
func requestGlobalMiddleware(r *gin.Engine) {
	r.Use(middleware.Cors())
}

// registerRoutes 注册各业务路由的具体处理函数
func registerRoutes(r *gin.Engine) {
	api := r.Group("/api")
	ws := r.Group("/ws")

	registerCommonRoutes(api.Group("/common"))
	chat.NewModule(chat.Deps{
		DB:            global.DB,
		RDB:           global.RDB,
		Log:           global.Log,
		AuthUser:      middleware.Authentication(jwtx.COMMON_USER),
		RateLimitUser: middleware.RateLimitUser,
		GetUserID:     jwtx.GetUserID,
		ParseToken: func(token string) (string, error) {
			claims, err := jwtx.ParseTokenString(token)
			if err != nil {
				return "", err
			}
			return claims.UserID, nil
		},
		Respond: response.Response,
	}).RegisterRoutes(api.Group("/chat"), ws)
	registerUserRoutes(api.Group("/user"))
	registerBabyRoutes(api.Group("/baby"))
	registerPostRoutes(api.Group("/post"))
	registerAdminRoutes(api.Group("/admin"))
	registerWSRoutes(ws)
}
