package smsx

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/global"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrSMSDisabled      = errors.New("短信服务未启用")
	ErrSMSConfigMissing = errors.New("短信配置缺失")
	ErrSMSSendFailed    = errors.New("短信发送失败")
)

//go:embed scripts/verify.lua
var verifyScript string

type SmsX struct {
	config config.SMS
	ttl    time.Duration
	rdb    redis.Cmdable
	client *http.Client
}

func NewSmsX() *SmsX {
	return &SmsX{
		config: config.Conf.SMS,
		ttl:    10 * time.Minute,
		rdb:    global.RDB,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (sx *SmsX) SendRegisterCode(ctx context.Context, phone string, code string) error {
	return sx.sendAndStore(ctx, fmt.Sprintf(constant.REGISTER_SMS_CODE_KEY, phone), phone, code)
}

func (sx *SmsX) SendBindPhoneCode(ctx context.Context, phone string, code string) error {
	return sx.sendAndStore(ctx, fmt.Sprintf(constant.BIND_PHONE_CODE_KEY, phone), phone, code)
}

func (sx *SmsX) SendRebindPhoneCode(ctx context.Context, phone string, code string) error {
	return sx.sendAndStore(ctx, fmt.Sprintf(constant.REBIND_PHONE_CODE_KEY, phone), phone, code)
}

func (sx *SmsX) sendAndStore(ctx context.Context, redisKey string, phone string, code string) error {
	if !sx.config.Enable {
		return ErrSMSDisabled
	}
	if sx.rdb == nil {
		return ErrSMSSendFailed
	}
	if err := sx.send(ctx, phone, code); err != nil {
		return err
	}
	return sx.rdb.Set(ctx, redisKey, code, sx.ttl).Err()
}

func (sx *SmsX) send(ctx context.Context, phone string, code string) error {
	if sx.config.Endpoint == "" {
		return ErrSMSConfigMissing
	}

	endpoint := sx.config.Endpoint
	if strings.Contains(endpoint, "%s") {
		if sx.config.Key == "" {
			return ErrSMSConfigMissing
		}
		endpoint = fmt.Sprintf(endpoint, sx.config.Key)
	}

	values := url.Values{}
	values.Set("targets", phone)
	values.Set("code", code)
	if sx.config.Name != "" {
		values.Set("name", sx.config.Name)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := sx.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("%w: status=%d body=%s", ErrSMSSendFailed, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

func (sx *SmsX) VerifyCode(ctx context.Context, key, code string) (bool, error) {
	if sx.rdb == nil {
		return false, nil
	}
	res, err := sx.rdb.Eval(ctx, verifyScript, []string{key}, code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if v, ok := res.(int64); ok {
		return v == 1, nil
	}
	return false, nil
}

func GenCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = b[i]%10 + '0'
	}
	return string(b)
}
