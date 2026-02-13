package manager

import (
	"nurture/internal/middleware"

	"github.com/gin-gonic/gin"
)

//主要管理路由组和中间件的注册

// PathHandler 是一个用于注册路由组的函数类型
type PathHandler func(rg *gin.RouterGroup)

// Middleware 是一个用于生成中间件的函数类型
type Middleware func() gin.HandlerFunc

// RouteManager 管理不同的路由组，按业务功能分组
type RouteManager struct {
	CommonRoutes *gin.RouterGroup //通用功能相关的路由组
	UserRoutes   *gin.RouterGroup //用户相关的路由组
	BabyRoutes   *gin.RouterGroup //宝宝相关的路由组
	PostRoutes   *gin.RouterGroup //帖子相关的路由组
	WSRoutes     *gin.RouterGroup //WebSocket相关的路由组
}

// NewRouteManager 创建一个新的 RouteManager 实例，包含各业务功能的路由组
func NewRouteManager(router *gin.Engine) *RouteManager {
	return &RouteManager{
		CommonRoutes: router.Group("/api/common"), //通用功能相关的路由组
		UserRoutes:   router.Group("/api/user"),   //用户相关的路由组
		BabyRoutes:   router.Group("/api/baby"),   //宝宝相关的路由组
		PostRoutes:   router.Group("/api/post"),   //帖子相关的路由组
		WSRoutes:     router.Group("/ws"),         //WebSocket相关的路由组
	}
}

// RegisterCommonRoutes通用功能相关的路由组
func (rm *RouteManager) RegisterCommonRoutes(handler PathHandler) {
	handler(rm.CommonRoutes)
}

// RegisterUserRoutes 用户相关的路由组
func (rm *RouteManager) RegisterUserRoutes(handler PathHandler) {
	handler(rm.UserRoutes)
}

// RegisterBabyRoutes 宝宝相关的路由组
func (rm *RouteManager) RegisterBabyRoutes(handler PathHandler) {
	handler(rm.BabyRoutes)
}

// RegisterPostRoutes 帖子相关的路由组
func (rm *RouteManager) RegisterPostRoutes(handler PathHandler) {
	handler(rm.PostRoutes)
}

// RegisterWSRoutes WebSocket相关的路由组
func (rm *RouteManager) RegisterWSRoutes(handler PathHandler) {
	handler(rm.WSRoutes)
}

// RegisterMiddleware 根据组名为对应的路由组注册中间件
func (rm *RouteManager) RegisterMiddleware(group string, middleware Middleware) {
	switch group {
	case "common":
		rm.CommonRoutes.Use(middleware())
	case "user":
		rm.UserRoutes.Use(middleware())
	case "baby":
		rm.BabyRoutes.Use(middleware())
	case "post":
		rm.PostRoutes.Use(middleware())
	case "ws":
		rm.WSRoutes.Use(middleware())
	default:
		// 处理未知的组名
		panic("unknown group name")
	}
}

// RequestGlobalMiddleware 注册全局中间件，应用于所有路由
func RequestGlobalMiddleware(r *gin.Engine) {
	r.Use(middleware.Cors())
}
