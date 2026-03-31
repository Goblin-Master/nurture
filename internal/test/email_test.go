package test

import (
	"context"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/pkg/emailx"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailSend(t *testing.T) {
	if os.Getenv("NURTURE_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("skip integration test: set NURTURE_RUN_INTEGRATION_TESTS=1 to run")
	}
	// 加载配置
	config.LoadConfig()
	// 初始化全局配置
	global.Init()

	// 初始化 EmailX
	ex := emailx.NewEmailX()

	err := ex.SendRegisterCode(context.Background(), config.Conf.Email.SendEmail, "123456")
	t.Logf("SendRegisterCode err: %v", err)
	assert.NoError(t, err)
}
