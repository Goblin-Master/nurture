package logic

import "errors"

var (
	ErrParamsType            = errors.New("参数格式错误")
	ErrDefault               = errors.New("默认错误")
	ErrBabyNotExist          = errors.New("宝宝不存在")
	ErrBabyGrowthNotExist    = errors.New("成长记录不存在")
	ErrVaccineRecordNotExist = errors.New("接种记录不存在")
	ErrInvalidVaccineStatus  = errors.New("疫苗状态非法")
	ErrInvalidActualTime     = errors.New("接种时间非法")
)
