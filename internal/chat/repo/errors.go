package repo

import "errors"

var (
	ErrDefault    = errors.New("默认错误")
	ErrParamsType = errors.New("参数格式错误")
)

var (
	ErrGroupNotExist          = errors.New("群不存在")
	ErrGroupFull              = errors.New("群人数已满")
	ErrNotMember              = errors.New("不是群成员")
	ErrPermissionDenied       = errors.New("无权限")
	ErrOwnerMustTransferFirst = errors.New("群主需要先转让再退出")
)
