package logic

import (
	"errors"
	"fmt"
	"nurture/internal/constant"
)

var (
	ErrParamsType            = errors.New("参数格式错误")
	ErrDefault               = errors.New("默认错误")
	ErrInvalidPhone          = errors.New("手机号格式错误")
	ErrInvalidBirthdayFormat = errors.New("生日格式错误")
	ErrInvalidGender         = errors.New("性别格式错误")
	ErrPartnerGenderMismatch = errors.New("双方性别不匹配")
	ErrPartnerAlreadyBound   = errors.New("已绑定另一半，不能重复绑定不同对象")
	ErrProfileUpdateFailed   = errors.New("资料更新失败")
	ErrFileOverSize          = fmt.Errorf("文件大小不能超过%dMB", constant.FILE_MAX_SIZE/1024/1024)
	ErrFileRead              = errors.New("文件读取失败")
	ErrFileUpload            = errors.New("文件上传失败")
)

var (
	ErrPostNotExist      = errors.New("帖子不存在")
	ErrInvalidPostStatus = errors.New("帖子状态非法")
)

var (
	ErrLoginWithFailedWay = errors.New("暂不支持这种登录方式")
	ErrAccountOrPassword  = errors.New("账号或密码错误")
	ErrEmail              = errors.New("邮箱错误")
	ErrCodeGet            = errors.New("code获取失败")
	ErrCodeVerify         = errors.New("验证码错误")
	ErrEmailIsUsed        = errors.New("邮箱已经被使用")
	ErrAccountIsUsed      = errors.New("账号已经被使用")
	ErrPhoneIsUsed        = errors.New("手机号已经被使用")
	ErrUserNotExist       = errors.New("用户不存在")
)

var (
	ErrPasswordEmpty    = errors.New("密码不能为空")
	ErrPasswordTooShort = errors.New("密码长度过短")
	ErrPasswordTooLong  = errors.New("密码长度过长")
	ErrPasswordTooWeak  = errors.New("密码强度过低")
)

var (
	ErrKnowledgeUpload  = errors.New("知识库上传失败")
	ErrChatStream       = errors.New("对话流失败")
	ErrInvalidSpaceType = errors.New("无效的知识空间类型")
	ErrLLMGenerate      = errors.New("LLM 生成失败")
	ErrEmbedding        = errors.New("向量化失败")
)

var (
	ErrBabyNotExist          = errors.New("宝宝不存在")
	ErrBabyGrowthNotExist    = errors.New("成长记录不存在")
	ErrVaccineRecordNotExist = errors.New("接种记录不存在")
	ErrInvalidVaccineStatus  = errors.New("疫苗状态非法")
	ErrInvalidActualTime     = errors.New("接种时间非法")
)
