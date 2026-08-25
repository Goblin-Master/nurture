package emailx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"nurture/internal/config"
	"nurture/internal/pkg/redisx"

	"github.com/stretchr/testify/assert"
)

func TestEmailXImplementsSender(t *testing.T) {
	var _ Sender = (*EmailX)(nil)
}

func TestSendRegisterCodeDisabled(t *testing.T) {
	ex := &EmailX{
		config: config.Email{Enable: false},
	}

	err := ex.SendCode(context.Background(), "to@example.com", "注册账号", "验证码是：123456", "test:email", "123456")
	if !errors.Is(err, ErrEmailDisabled) {
		t.Fatalf("SendCode() error = %v, want ErrEmailDisabled", err)
	}
}

func TestVerifyCodeWithoutRedisReturnsFalse(t *testing.T) {
	ex := &EmailX{}

	ok, err := ex.VerifyCode(context.Background(), "test:email", "123456")

	if err != nil {
		t.Fatalf("VerifyCode() error = %v, want nil", err)
	}
	if ok {
		t.Fatal("VerifyCode() = true, want false without redis")
	}
}

func TestGenCodeReturnsSixDigits(t *testing.T) {
	code, err := GenCode()

	if err != nil {
		t.Fatalf("GenCode() error = %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("GenCode() length = %d, want 6", len(code))
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			t.Fatalf("GenCode() = %q, want digits only", code)
		}
	}
}

func TestGenCodeReturnsRandomReaderError(t *testing.T) {
	oldReader := codeRandReader
	codeRandReader = failingReader{}
	t.Cleanup(func() {
		codeRandReader = oldReader
	})

	code, err := GenCode()

	if err == nil {
		t.Fatal("GenCode() error = nil, want error")
	}
	if code != "" {
		t.Fatalf("GenCode() code = %q, want empty on error", code)
	}
}

func TestVerifyScriptEmbedded(t *testing.T) {
	if !strings.Contains(verifyScript, `redis.call("DEL", KEYS[1])`) {
		t.Fatalf("verifyScript was not embedded from scripts/verify.lua")
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

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("random failed")
}
