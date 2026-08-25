package test

import (
	"testing"

	"nurture/internal/post"

	"github.com/gin-gonic/gin"
)

func TestPostRoutesUseResourcePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	post.NewModule(post.Deps{}).RegisterRoutes(r.Group("/api/post"))

	routes := routeSet(r.Routes())
	for _, route := range []string{
		"POST /api/post",
		"GET /api/post/tags/:tag_id/posts",
		"GET /api/post/mine/milestones",
		"POST /api/post/drafts/:post_id/publish",
		"PUT /api/post/drafts/:post_id",
		"DELETE /api/post/drafts/:post_id",
		"DELETE /api/post/:post_id",
	} {
		if !routes[route] {
			t.Fatalf("route %s is not registered", route)
		}
	}
	for _, route := range []string{
		"POST /api/post/newPost",
		"GET /api/post/tag/:tag_id",
		"GET /api/post/mine/milestone",
		"POST /api/post/:post_id/publish",
		"PUT /api/post/:post_id",
		"DELETE /api/post/:post_id/delete",
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
