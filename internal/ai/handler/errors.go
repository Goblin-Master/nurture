package handler

import "errors"

var (
	ErrSessionIDEmpty   = errors.New("会话ID不能为空")
	ErrPermissionDenied = errors.New("权限不足")
)
