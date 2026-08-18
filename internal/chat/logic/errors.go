package logic

import (
	"errors"
	"nurture/internal/chat/repo"
)

var (
	ErrParamsType = errors.New("参数格式错误")
	ErrDefault    = errors.New("默认错误")
)

var (
	ErrTooManyRequests    = errors.New("请求过于频繁")
	ErrGroupNotExist      = errors.New("群不存在")
	ErrGroupFull          = errors.New("群人数已满")
	ErrNotMember          = errors.New("不是群成员")
	ErrPermissionDenied   = errors.New("无权限")
	ErrOwnerTransferFirst = errors.New("群主需要先转让再退出")
	ErrInvalidCursor      = errors.New("游标格式错误")
	ErrInvalidMessageType = errors.New("消息类型非法")
)

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, repo.ErrParamsType):
		return ErrParamsType
	case errors.Is(err, repo.ErrGroupNotExist):
		return ErrGroupNotExist
	case errors.Is(err, repo.ErrGroupFull):
		return ErrGroupFull
	case errors.Is(err, repo.ErrNotMember):
		return ErrNotMember
	case errors.Is(err, repo.ErrPermissionDenied):
		return ErrPermissionDenied
	case errors.Is(err, repo.ErrOwnerMustTransferFirst):
		return ErrOwnerTransferFirst
	default:
		return ErrDefault
	}
}
