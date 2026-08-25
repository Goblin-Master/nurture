package repo

import "errors"

var (
	ErrDocumentAdd    = errors.New("文档添加失败")
	ErrDocumentSearch = errors.New("文档检索失败")
	ErrHistoryGet     = errors.New("获取对话历史失败")
	ErrHistorySave    = errors.New("保存对话历史失败")
)
