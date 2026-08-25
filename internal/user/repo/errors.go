package repo

import "errors"

var (
	ErrDefault          = errors.New("默认错误")
	ErrParamsType       = errors.New("参数格式错误")
	ErrUserUpdateFailed = errors.New("用户资料更新失败")
	ErrUserNotExist     = errors.New("用户不存在")
	ErrAccountIsUsed    = errors.New("账号已经被使用")
	ErrEmailIsUsed      = errors.New("邮箱已经被使用")
	ErrPhoneIsUsed      = errors.New("手机号已经被使用")
	ErrAccountOrPwd     = errors.New("账号或密码错误")
)
