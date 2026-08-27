package logic

import "errors"

var (
	ErrParamsType            = errors.New("参数格式错误")
	ErrDefault               = errors.New("默认错误")
	ErrInvalidPhone          = errors.New("手机号格式错误")
	ErrInvalidBirthdayFormat = errors.New("生日格式错误")
	ErrInvalidGender         = errors.New("性别格式错误")
	ErrPartnerGenderMismatch = errors.New("双方性别不匹配")
	ErrPartnerAlreadyBound   = errors.New("已绑定另一半，不能重复绑定不同对象")
	ErrProfileUpdateFailed   = errors.New("资料更新失败")
)

var (
	ErrLoginWithFailedWay = errors.New("暂不支持这种登录方式")
	ErrAccountOrPassword  = errors.New("账号或密码错误")
	ErrEmail              = errors.New("邮箱错误")
	ErrCodeGet            = errors.New("code获取失败")
	ErrCodeVerify         = errors.New("验证码错误")
	ErrEmailIsUsed        = errors.New("邮箱已经被使用")
	ErrAccountIsUsed      = errors.New("账号已经被使用")
	ErrPhoneIsUsed        = errors.New("手机号已经被使用")
	ErrUserNotExist       = errors.New("用户不存在")
	ErrAuthUnavailable    = errors.New("认证服务不可用")
	ErrRTokenInvalid      = errors.New("刷新凭证无效")
	ErrRTokenReplay       = errors.New("刷新凭证已失效，请重新登录")
)

var (
	ErrPasswordEmpty    = errors.New("密码不能为空")
	ErrPasswordTooShort = errors.New("密码长度过短")
	ErrPasswordTooLong  = errors.New("密码长度过长")
	ErrPasswordTooWeak  = errors.New("密码强度过低")
)
