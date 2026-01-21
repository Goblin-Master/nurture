package test

import (
	"context"
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/pkg/emailx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailSend(t *testing.T) {
	// 加载配置
	config.LoadConfig()
	// 初始化日志
	global.Init()

	// 初始化 EmailX
	ex := emailx.NewEmailX()

	err := ex.SendRegisterCode(context.Background(), config.Conf.Email.SendEmail, "123456")
	t.Logf("SendRegisterCode err: %v", err)
	assert.NoError(t, err)
}
