package chat

import "errors"

var (
	errTokenEmpty   = errors.New("token不能为空")
	errTokenInvalid = errors.New("token无效")
)
