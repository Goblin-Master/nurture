package emailx

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"net/smtp"
	"nurture/internal/config"
	"nurture/internal/global"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jordan-wright/email"
)

var (
	ErrEmailDisabled = errors.New("邮件服务未启用")
	ErrEmailStore    = errors.New("邮件验证码保存失败")
	ErrSendOverTime  = errors.New("邮件发送超时")
)

//go:embed scripts/verify.lua
var verifyScript string

type EmailX struct {
	config config.Email
	ttl    time.Duration
	rdb    redis.Cmdable
}

func NewEmailX() *EmailX {
	return &EmailX{
		config: config.Conf.Email,
		ttl:    10 * time.Minute,
		rdb:    global.RDB,
	}
}

func (ex *EmailX) SendCode(ctx context.Context, to string, title string, text string, key string, code string) error {
	subject := fmt.Sprintf("[%s]%s", ex.config.Subject, title)
	if err := ex.sendEmail(ctx, to, subject, text); err != nil {
		return err
	}
	return ex.storeCode(ctx, key, code)
}

func (ex *EmailX) storeCode(ctx context.Context, key string, code string) error {
	if ex.rdb == nil {
		return ErrEmailStore
	}
	return ex.rdb.Set(ctx, key, code, ex.ttl).Err()
}

func (ex *EmailX) sendEmail(ctx context.Context, to, subject, text string) error {
	if !ex.config.Enable {
		return ErrEmailDisabled
	}
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", ex.config.SendNickname, ex.config.SendEmail)
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(text)

	addr := fmt.Sprintf("%s:%d", ex.config.Domain, ex.config.Port)
	auth := smtp.PlainAuth("", ex.config.SendEmail, ex.config.AuthCode, ex.config.Domain)

	type result struct{ err error }
	done := make(chan result, 1)

	// 1. 计算剩余时间
	var timeout time.Duration
	if dl, ok := ctx.Deadline(); ok {
		timeout = time.Until(dl)
		if timeout <= 0 {
			return context.DeadlineExceeded
		}
	} else {
		timeout = 10 * time.Second // 调用方没给 deadline 就用默认,10s 是为了防止邮件发送失败
	}

	// 2. 异步发送
	go func() {
		err := e.Send(addr, auth)
		// 过滤掉某些老旧服务器返回的“short response”伪错误
		if err != nil && !strings.Contains(err.Error(), "short response") {
			done <- result{err: err}
			return
		}
		done <- result{err: nil}
	}()

	// 3. 等待完成或超时 / 被取消
	select {
	case res := <-done:
		return res.err
	case <-time.After(timeout):
		return ErrSendOverTime
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ex *EmailX) VerifyCode(ctx context.Context, key, code string) (bool, error) {
	res, err := ex.rdb.Eval(ctx, verifyScript, []string{key}, code).Result()
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
