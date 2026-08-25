package logic

import "errors"

var (
	ErrParamsType          = errors.New("参数格式错误")
	ErrDefault             = errors.New("默认错误")
	ErrDatabaseUnavailable = errors.New("数据库服务未启用")
)

var (
	ErrKnowledgeUpload  = errors.New("知识库上传失败")
	ErrChatStream       = errors.New("对话流失败")
	ErrInvalidSpaceType = errors.New("无效的知识空间类型")
)

var (
	ErrBabyNotExist = errors.New("宝宝不存在")
)
