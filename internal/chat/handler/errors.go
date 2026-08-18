package handler

import "errors"

var (
	ErrTokenEmpty   = errors.New("token不能为空")
	ErrTokenInvalid = errors.New("token无效")
)
