package router

import (
	"nurture/internal/dto"
	"nurture/internal/handler"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(rg *gin.RouterGroup) {
	userHandler := handler.NewUserHandler()

	rg.GET("/users/list",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindQueryMiddleware[dto.AdminListUsersReq],
		userHandler.AdminListUsers,
	)
	rg.PUT("/users/:user_id/role/admin",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindUriMiddleware[dto.AdminPromoteUri],
		userHandler.AdminPromoteToAdmin,
	)
}
