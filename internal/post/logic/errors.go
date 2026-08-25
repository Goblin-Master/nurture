package logic

import "errors"

var (
	ErrParamsType        = errors.New("参数格式错误")
	ErrDefault           = errors.New("默认错误")
	ErrPostNotExist      = errors.New("帖子不存在")
	ErrInvalidPostStatus = errors.New("帖子状态非法")
)
