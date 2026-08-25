package user

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	userHandler := m.handler

	rg.POST("/login", middleware.BindJsonMiddleware[dto.LoginReq], userHandler.Login)
	rg.POST("/register", middleware.BindJsonMiddleware[dto.RegisterReq], userHandler.Register)
	rg.POST("/register/sms", middleware.BindJsonMiddleware[dto.RegisterSMSReq], userHandler.RegisterSMS)
	rg.POST("/code/login", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetLoginCode)
	rg.POST("/code/register", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetRegisterCode)
	rg.POST("/code/register/sms", middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetRegisterSMSCode)
	rg.POST("/code/reset", middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetResetCode)
	rg.POST("/resetPassword", middleware.BindJsonMiddleware[dto.ResetPasswordReq], userHandler.ResetPassword)
	rg.PUT("/profile", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.UpdateUserAdditionReq], userHandler.UpdateProfile)
	rg.PUT("/avatar", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.UpdateAvatarReq], userHandler.UpdateAvatar)
	rg.POST("/code/bind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetBindPhoneCode)
	rg.POST("/bind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindPhoneReq], userHandler.BindPhone)
	rg.POST("/code/bind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetBindEmailCode)
	rg.POST("/bind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindEmailReq], userHandler.BindEmail)
	rg.POST("/code/rebind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetSMSCodeReq], userHandler.GetRebindPhoneCode)
	rg.POST("/rebind/phone", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindPhoneReq], userHandler.RebindPhone)
	rg.POST("/code/rebind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.GetCodeReq], userHandler.GetRebindEmailCode)
	rg.POST("/rebind/email", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.BindEmailReq], userHandler.RebindEmail)
	rg.GET("/me", middleware.Authentication(jwtx.COMMON_USER), userHandler.MyProfile)
	rg.POST("/partner/bind", middleware.Authentication(jwtx.COMMON_USER), middleware.BindJsonMiddleware[dto.PartnerBindReq], userHandler.BindPartner)
	rg.GET("/partner", middleware.Authentication(jwtx.COMMON_USER), userHandler.GetPartner)
	rg.POST("/follow/:target_user_id",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindUriMiddleware[dto.FollowReq],
		userHandler.Follow,
	)
	rg.DELETE("/follow/:target_user_id",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindUriMiddleware[dto.FollowReq],
		userHandler.Unfollow,
	)
	rg.GET("/following",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindQueryMiddleware[dto.FollowingListReq],
		userHandler.ListFollowing,
	)
	rg.GET("/followers",
		middleware.Authentication(jwtx.COMMON_USER),
		middleware.BindQueryMiddleware[dto.FollowersListReq],
		userHandler.ListFollowers,
	)
}

func (m *Module) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/users/list",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindQueryMiddleware[dto.AdminListUsersReq],
		m.handler.AdminListUsers,
	)
	rg.PUT("/users/:user_id/role/admin",
		middleware.Authentication(jwtx.ADMIN),
		middleware.BindUriMiddleware[dto.AdminPromoteUri],
		m.handler.AdminPromoteToAdmin,
	)
}
