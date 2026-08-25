package repo

import "errors"

var (
	ErrDefault           = errors.New("默认错误")
	ErrParamsType        = errors.New("参数格式错误")
	ErrPostNotExist      = errors.New("帖子不存在")
	ErrInvalidPostStatus = errors.New("帖子状态非法")
	ErrPostNotDraft      = errors.New("帖子不是草稿，无法发布")
)
