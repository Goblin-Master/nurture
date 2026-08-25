package repo

import "errors"

var (
	ErrDefault             = errors.New("默认错误")
	ErrBabyNotExist        = errors.New("宝宝不存在")
	ErrBabyGrowthNotExist  = errors.New("宝宝成长记录不存在")
	ErrBabyVaccineNotExist = errors.New("宝宝疫苗记录不存在")
)
