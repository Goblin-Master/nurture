package test

import (
	"testing"

	"nurture/internal/user"

	"github.com/gin-gonic/gin"
)

func TestUserRoutesUseResourcePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	module := user.NewModule(user.Deps{})
	module.RegisterRoutes(r.Group("/api/user"))
	module.RegisterAdminRoutes(r.Group("/api/admin"))

	routes := routeSet(r.Routes())
	for _, route := range []string{
		"POST /api/user/password/reset",
		"POST /api/user/partner",
		"POST /api/user/following/:target_user_id",
		"DELETE /api/user/following/:target_user_id",
		"GET /api/admin/users",
	} {
		if !routes[route] {
			t.Fatalf("route %s is not registered", route)
		}
	}
	for _, route := range []string{
		"POST /api/user/resetPassword",
		"POST /api/user/partner/bind",
		"POST /api/user/follow/:target_user_id",
		"DELETE /api/user/follow/:target_user_id",
		"GET /api/admin/users/list",
	} {
		if routes[route] {
			t.Fatalf("legacy route %s should not be registered", route)
		}
	}
}

func routeSet(routes gin.RoutesInfo) map[string]bool {
	ret := make(map[string]bool, len(routes))
	for _, route := range routes {
		ret[route.Method+" "+route.Path] = true
	}
	return ret
}
