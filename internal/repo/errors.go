package repo

import "errors"

var (
	ErrDefault          = errors.New("默认错误")
	ErrUserUpdateFailed = errors.New("用户资料更新失败")
	ErrVectorStoreInit  = errors.New("向量存储初始化失败")
	ErrDocumentAdd      = errors.New("文档添加失败")
	ErrDocumentSearch   = errors.New("文档检索失败")
	ErrHistoryGet       = errors.New("获取对话历史失败")
	ErrHistorySave      = errors.New("保存对话历史失败")
	ErrParamsType       = errors.New("参数格式错误")

	// User related errors restored
	ErrUserNotExist  = errors.New("用户不存在")
	ErrAccountIsUsed = errors.New("账号已经被使用")
	ErrEmailIsUsed   = errors.New("邮箱已经被使用")

	// Baby related errors
	ErrBabyNotExist        = errors.New("宝宝不存在")
	ErrBabyGrowthNotExist  = errors.New("宝宝成长记录不存在")
	ErrBabyVaccineNotExist = errors.New("宝宝疫苗记录不存在")

	// Post related errors
	ErrPostNotExist      = errors.New("帖子不存在")
	ErrInvalidPostStatus = errors.New("帖子状态非法")
	ErrPostNotDraft      = errors.New("帖子不是草稿，无法发布")
)
