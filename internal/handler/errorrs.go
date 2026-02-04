package handler

import "errors"

var (
	ErrChatStream       = errors.New("对话流失败")
	ErrFileRead         = errors.New("文件读取失败")
	ErrSessionIDEmpty   = errors.New("会话ID不能为空")
	ErrPermissionDenied = errors.New("权限不足")
)
