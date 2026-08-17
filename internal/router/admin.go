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
	postHandler := handler.NewPostHandler()
	babyHandler := handler.NewBabyHandler()

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
	rg.POST("/tags",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindJsonMiddleware[dto.AdminTagCreateReq],
		postHandler.AdminCreateTag,
	)
	rg.DELETE("/tags/:tag_id",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindUriMiddleware[dto.AdminTagDeleteUri],
		postHandler.AdminDeleteTag,
	)
	rg.POST("/vaccines",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindJsonMiddleware[dto.AdminCreateVaccineReq],
		babyHandler.AdminCreateVaccine,
	)
}
