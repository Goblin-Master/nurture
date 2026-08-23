package emailx

import (
	"context"
	"nurture/internal/config"
	"nurture/internal/pkg/redisx"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmailSend(t *testing.T) {
	if os.Getenv("NURTURE_RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("skip integration test: set NURTURE_RUN_INTEGRATION_TESTS=1 to run")
	}
	// 加载配置
	config.LoadConfig()
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

	err := ex.SendRegisterCode(context.Background(), config.Conf.Email.SendEmail, "123456")
	t.Logf("SendRegisterCode err: %v", err)
	assert.NoError(t, err)
}
