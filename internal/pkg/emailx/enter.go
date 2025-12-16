package emailx

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/smtp"
	"nurture/internal/config"
	"nurture/internal/constant"
	"nurture/internal/global"
	"nurture/internal/pkg/syncx"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

var (
	ErrSendOverTime = errors.New("邮件发送超时")
)

type EmailX struct {
	config config.Email
	ttl    time.Duration
	store  *syncx.Map[string, string]
}

func NewEmailX() *EmailX {
	return &EmailX{
		config: config.Conf.Email,
		ttl:    10 * time.Minute,
		store:  global.CodeStore,
	}
}

func (ex *EmailX) SendLoginCode(ctx context.Context, to string, code string) (err error) {
	subject := fmt.Sprintf("[%s]邮箱登录", ex.config.Subject)
	text := fmt.Sprintf("你正在进行邮箱登录，登录的验证码是：%s，十分钟内有效", code)
	if err := ex.sendEmail(ctx, to, subject, text); err != nil {
		return err
	}
	ex.store.StoreWithTTL(fmt.Sprintf(constant.LOGIN_CODE_KEY, to), code, ex.ttl)
	return nil
}
func (ex *EmailX) SendResetPwdCode(ctx context.Context, to string, code string) (err error) {
	subject := fmt.Sprintf("[%s]重置密码", ex.config.Subject)
	text := fmt.Sprintf("你正在进行账号密码重置，重置的验证码是：%s，十分钟内有效", code)
	if err := ex.sendEmail(ctx, to, subject, text); err != nil {
		return err
	}
	ex.store.StoreWithTTL(fmt.Sprintf(constant.RESET_PWD_CODE_KEY, to), code, ex.ttl)
	return nil
}

func (ex *EmailX) SendRegisterCode(ctx context.Context, to string, code string) (err error) {
	subject := fmt.Sprintf("[%s]注册账号", ex.config.Subject)
	text := fmt.Sprintf("你正在进行账号注册，注册的验证码是：%s，十分钟内有效", code)
	if err := ex.sendEmail(ctx, to, subject, text); err != nil {
		return err
	}
	ex.store.StoreWithTTL(fmt.Sprintf(constant.REGISTER_CODE_KEY, to), code, ex.ttl)
	return nil
}

func (ex *EmailX) sendEmail(ctx context.Context, to, subject, text string) error {
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
		timeout = 3 * time.Second // 调用方没给 deadline 就用默认
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

func (ex *EmailX) VerifyCode(key, code string) bool {
	var ans bool
	if v, ok := ex.store.Load(key); ok {
		if v == code {
			ans = true
			ex.store.Delete(key)
		}
	}
	return ans
}

func (ex *EmailX) ShowDataForDebug() {
	ex.store.Range(func(key, value string) bool {
		global.Log.Debugf("key:%s, value:%s", key, value)
		return true
	})
}

func GenCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = b[i]%10 + '0'
	}
	return string(b)
}
