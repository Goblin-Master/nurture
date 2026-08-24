package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/pkg/response"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListenHealthzWithDBDisabled(t *testing.T) {
	oldMode := gin.Mode()
	oldApp := config.Conf.App
	oldDBConfig := config.Conf.DB
	oldDB := global.DB
	oldRDB := global.RDB
	oldRMQ := global.RMQ

	gin.SetMode(gin.TestMode)
	config.Conf.App = config.App{Env: "dev"}
	config.Conf.DB = config.DB{Enable: false}
	global.DB = nil
	global.RDB = nil
	global.RMQ = nil
	t.Cleanup(func() {
		gin.SetMode(oldMode)
		config.Conf.App = oldApp
		config.Conf.DB = oldDBConfig
		global.DB = oldDB
		global.RDB = oldRDB
		global.RMQ = oldRMQ
	})

	r, err := listen()
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want %d", w.Code, http.StatusOK)
	}

	var body response.Body
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode healthz body: %v", err)
	}
	if body.Code != 0 || body.Message != "OK" || body.Data != "ok" {
		t.Fatalf("healthz body = %+v", body)
	}
}

func TestDBRoutesReturnUnavailableWhenDBDisabled(t *testing.T) {
	oldMode := gin.Mode()
	oldDBConfig := config.Conf.DB
	oldDB := global.DB
	oldRDB := global.RDB
	oldRMQ := global.RMQ

	gin.SetMode(gin.TestMode)
	config.Conf.DB = config.DB{Enable: false}
	global.DB = nil
	global.RDB = nil
	global.RMQ = nil
	t.Cleanup(func() {
		gin.SetMode(oldMode)
		config.Conf.DB = oldDBConfig
		global.DB = oldDB
		global.RDB = oldRDB
		global.RMQ = oldRMQ
	})

	r, err := listen()
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("db route status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var body response.Body
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode db route body: %v", err)
	}
	if body.Code != -1 || body.Message != "数据库服务未启用" {
		t.Fatalf("db route body = %+v", body)
	}
}
