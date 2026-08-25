package repo

import "errors"

var (
	ErrDefault         = errors.New("默认错误")
	ErrVectorStoreInit = errors.New("向量存储初始化失败")
	ErrDocumentAdd     = errors.New("文档添加失败")
	ErrDocumentSearch  = errors.New("文档检索失败")
	ErrHistoryGet      = errors.New("获取对话历史失败")
	ErrHistorySave     = errors.New("保存对话历史失败")
	ErrParamsType      = errors.New("参数格式错误")
)
