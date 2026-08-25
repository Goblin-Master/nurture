package smsx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"nurture/internal/config"
)

func TestSmsXImplementsSender(t *testing.T) {
	var _ Sender = (*SmsX)(nil)
}

func TestSendRegisterCodeDisabled(t *testing.T) {
	sx := &SmsX{
		config: config.SMS{Enable: false},
	}

	err := sx.SendCode(context.Background(), "test:sms", "13800138000", "123456")
	if !errors.Is(err, ErrSMSDisabled) {
		t.Fatalf("SendCode() error = %v, want ErrSMSDisabled", err)
	}
}

func TestVerifyScriptEmbedded(t *testing.T) {
	if !strings.Contains(verifyScript, `redis.call("DEL", KEYS[1])`) {
		t.Fatalf("verifyScript was not embedded from scripts/verify.lua")
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

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("random failed")
}
