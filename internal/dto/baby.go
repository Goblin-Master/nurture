package dto

type (
	BabyItem struct {
		BabyID string `json:"baby_id"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	ChangeBabyReq  struct{}
	ChangeBabyResp struct {
		Babies []BabyItem `json:"babies"`
	}
)

type (
	NewBabyReq struct {
		Name              string  `json:"name" binding:"required"`
		Gender            string  `json:"gender" binding:"required"`
		Birthday          int64   `json:"birthday" binding:"required"`
		Avatar            string  `json:"avatar"`
		Height            float64 `json:"height"`
		Weight            float64 `json:"weight"`
		HeadCircumference float64 `json:"head_circumference"`
		Remark            string  `json:"remark"`
	}
	NewBabyResp struct {
		BabyID  string `json:"baby_id"`
		Message string `json:"message"`
	}
)

type (
	BabyProfileReq struct {
		BabyID string `json:"baby_id" form:"baby_id" binding:"required"`
	}
	BabyProfileResp struct {
		BabyID            string  `json:"baby_id"`
		Name              string  `json:"name"`
		Avatar            string  `json:"avatar"`
		Gender            string  `json:"gender"`
		Birthday          int64   `json:"birthday"`
		RecordTime        int64   `json:"record_time"`
		Height            float64 `json:"height"`
		Weight            float64 `json:"weight"`
		HeadCircumference float64 `json:"head_circumference"`
	}
)

type (
	GetVaccineListReq struct {
		BabyID string `json:"baby_id" form:"baby_id" binding:"required"`
	}
	VaccineItem struct {
		VaccineID  string `json:"vaccine_id"`
		Name       string `json:"name"`
		Disease    string `json:"disease"`
		DueTime    int64  `json:"due_time"`
		Status     string `json:"status"`
		ActualTime int64  `json:"actual_time"`
	}
	GetVaccineListResp struct {
		Items []VaccineItem `json:"items"`
	}
)

type (
	ChangeVaccineStatusReq struct {
		BabyID     string `json:"baby_id" binding:"required"`
		VaccineID  string `json:"vaccine_id" binding:"required"`
		Status     string `json:"status" binding:"required"`
		ActualTime int64  `json:"actual_time"`
	}
	ChangeVaccineStatusResp struct {
		Message string `json:"message"`
	}
)
