package logic

import (
	"errors"
	"fmt"
	"nurture/internal/constant"
)

var (
	ErrParamsType          = errors.New("参数格式错误")
	ErrDefault             = errors.New("默认错误")
	ErrFileOverSize        = fmt.Errorf("文件大小不能超过%dMB", constant.FILE_MAX_SIZE/1024/1024)
	ErrFileRead            = errors.New("文件读取失败")
	ErrFileUpload          = errors.New("文件上传失败")
	ErrDatabaseUnavailable = errors.New("数据库服务未启用")
)

var (
	ErrKnowledgeUpload  = errors.New("知识库上传失败")
	ErrChatStream       = errors.New("对话流失败")
	ErrInvalidSpaceType = errors.New("无效的知识空间类型")
	ErrLLMGenerate      = errors.New("LLM 生成失败")
	ErrEmbedding        = errors.New("向量化失败")
	ErrNotImplemented   = errors.New("功能未实现")
)

var (
	ErrBabyNotExist          = errors.New("宝宝不存在")
	ErrBabyGrowthNotExist    = errors.New("成长记录不存在")
	ErrVaccineRecordNotExist = errors.New("接种记录不存在")
	ErrInvalidVaccineStatus  = errors.New("疫苗状态非法")
	ErrInvalidActualTime     = errors.New("接种时间非法")
)
