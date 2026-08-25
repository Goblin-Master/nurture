package test

import (
	"testing"

	"nurture/internal/baby"

	"github.com/gin-gonic/gin"
)

func TestBabyRoutesUseResourcePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	baby.NewModule(baby.Deps{}).RegisterRoutes(r.Group("/api/baby"))

	routes := routeSet(r.Routes())
	for _, route := range []string{
		"GET /api/baby",
		"POST /api/baby",
		"POST /api/baby/growth-records",
		"GET /api/baby/growth-record",
		"GET /api/baby/growth-curve",
		"GET /api/baby/vaccines",
		"PUT /api/baby/vaccines/status",
		"POST /api/baby/photos",
		"DELETE /api/baby/photos",
		"GET /api/baby/photos",
	} {
		if !routes[route] {
			t.Fatalf("route %s is not registered", route)
		}
	}
	for _, route := range []string{
		"GET /api/baby/changeBaby",
		"POST /api/baby/newBaby",
		"POST /api/baby/growthRecords",
		"GET /api/baby/growthRecord",
		"GET /api/baby/growthCurve",
		"GET /api/baby/vaccine/getVaccineList",
		"PUT /api/baby/vaccine/changeStatus",
		"POST /api/baby/photo/upload",
		"DELETE /api/baby/photo/delete",
		"GET /api/baby/photo/list",
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
