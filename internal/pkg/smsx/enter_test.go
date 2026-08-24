package smsx

import (
	"context"
	"errors"
	"nurture/internal/config"
	"strings"
	"testing"
)

func TestSendRegisterCodeDisabled(t *testing.T) {
	sx := &SmsX{
		config: config.SMS{Enable: false},
	}

	err := sx.SendRegisterCode(context.Background(), "13800138000", "123456")
	if !errors.Is(err, ErrSMSDisabled) {
		t.Fatalf("SendRegisterCode() error = %v, want ErrSMSDisabled", err)
	}
}

func TestVerifyScriptEmbedded(t *testing.T) {
	if !strings.Contains(verifyScript, `redis.call("DEL", KEYS[1])`) {
		t.Fatalf("verifyScript was not embedded from scripts/verify.lua")
	}
}
