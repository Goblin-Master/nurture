package emailx

import (
	"context"
	"errors"
	"nurture/internal/config"
	"nurture/internal/pkg/redisx"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendRegisterCodeDisabled(t *testing.T) {
	ex := &EmailX{
		config: config.Email{Enable: false},
	}

	err := ex.SendCode(context.Background(), "to@example.com", "注册账号", "验证码是：123456", "test:email", "123456")
	if !errors.Is(err, ErrEmailDisabled) {
		t.Fatalf("SendCode() error = %v, want ErrEmailDisabled", err)
	}
}

func TestEmailSend(t *testing.T) {
	// 加载配置
	config.LoadConfig()
	if !config.Conf.Email.Enable {
		t.Skip("skip email integration test: email.enable=false")
	}
	if !config.Conf.Redis.Enable {
		t.Skip("skip email integration test: redis.enable=false")
	}
	rdb := redisx.InitRedis()
	if rdb == nil {
		t.Fatal("redis is disabled")
	}
	if closer, ok := rdb.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	// 初始化 EmailX
	ex := &EmailX{
		config: config.Conf.Email,
		ttl:    10 * time.Minute,
		rdb:    rdb,
	}

	err := ex.SendCode(context.Background(), config.Conf.Email.SendEmail, "注册账号", "验证码是：123456", "test:email", "123456")
	t.Logf("SendCode err: %v", err)
	assert.NoError(t, err)
}
