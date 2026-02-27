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

// 喂养记录（baby/daily）
type (
	FeedingCreateUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	FeedingCreateReq struct {
		FeedType string `json:"feed_type" binding:"required"`
		FeedTime int64  `json:"feed_time" binding:"required"`
		Remark   string `json:"remark"`
	}
	FeedingCreateResp struct {
		FeedingID string `json:"feeding_id"`
		Message   string `json:"message"`
	}
)

type (
	FeedingUpdateUri struct {
		BabyID    string `uri:"baby_id" binding:"required"`
		FeedingID string `uri:"feeding_id" binding:"required"`
	}
	FeedingUpdateReq struct {
		FeedType string `json:"feed_type" binding:"required"`
		FeedTime int64  `json:"feed_time" binding:"required"`
		Remark   string `json:"remark"`
	}
	FeedingUpdateResp struct {
		Message string `json:"message"`
	}
)

type (
	FeedingListAtUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	FeedingListAtReq struct {
		Date string `form:"date" binding:"required"`
	}
	FeedingItem struct {
		FeedingID string `json:"feeding_id"`
		FeedTime  int64  `json:"feed_time"`
		FeedType  string `json:"feed_type"`
		Remark    string `json:"remark"`
	}
	FeedingListAtResp struct {
		Items []FeedingItem `json:"items"`
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
	CreateGrowthReq struct {
		BabyID            string  `json:"baby_id" binding:"required"`
		RecordTime        int64   `json:"record_time" binding:"required"`
		Height            float64 `json:"height"`
		Weight            float64 `json:"weight"`
		HeadCircumference float64 `json:"head_circumference"`
		Remark            string  `json:"remark"`
	}
	CreateGrowthResp struct {
		RecordID string `json:"record_id"`
		Message  string `json:"message"`
	}
)

type (
	GetVaccineListReq struct {
		BabyID string `json:"baby_id" form:"baby_id" binding:"required"`
		Status string `json:"status" form:"status"`
	}
	VaccineItem struct {
		DoseID     string `json:"dose_id"`
		VaccineID  string `json:"vaccine_id"`
		Name       string `json:"name"`
		Disease    string `json:"disease"`
		Link       string `json:"link"`
		DoseNumber int32  `json:"dose_number"`
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
		DoseID     string `json:"dose_id" binding:"required"`
		Status     string `json:"status" binding:"required"`
		ActualTime int64  `json:"actual_time"`
	}
	ChangeVaccineStatusResp struct {
		Message string `json:"message"`
	}
)

type (
	UploadBabyPhotosReq struct {
		BabyID string   `json:"baby_id" binding:"required"`
		Links  []string `json:"links" binding:"required"`
	}

	PhotoItem struct {
		PhotoID string `json:"photo_id"`
		Link    string `json:"link"`
		Ctime   int64  `json:"ctime"`
	}

	UploadBabyPhotosResp struct {
		Items []PhotoItem `json:"items"`
	}

	DeleteBabyPhotosReq struct {
		BabyID   string   `json:"baby_id" binding:"required"`
		PhotoIDs []string `json:"photo_ids" binding:"required"`
	}

	DeleteBabyPhotosResp struct {
		Deleted int64 `json:"deleted"`
	}

	ListBabyPhotosReq struct {
		BabyID   string `form:"baby_id" binding:"required"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}

	ListBabyPhotosResp struct {
		Items    []PhotoItem `json:"items"`
		Page     int         `json:"page"`
		PageSize int         `json:"page_size"`
		HasMore  bool        `json:"has_more"`
	}
)

type (
	GrowthAtReq struct {
		BabyID string `json:"baby_id" form:"baby_id" binding:"required"`
		Time   int64  `json:"time" form:"time" binding:"required"`
	}
	GrowthAtResp struct {
		RecordID          string   `json:"record_id"`
		RecordTime        int64    `json:"record_time"`
		Height            *float64 `json:"height"`
		Weight            *float64 `json:"weight"`
		HeadCircumference *float64 `json:"head_circumference"`
		Remark            string   `json:"remark"`
		CreatedBy         string   `json:"created_by"`
		UpdatedBy         string   `json:"updated_by"`
	}
)

type (
	GrowthCurveReq struct {
		BabyID    string `form:"baby_id" binding:"required"`
		Metric    string `form:"metric" binding:"required"`
		From      int64  `form:"from"`
		To        int64  `form:"to"`
		MaxPoints int    `form:"max_points"`
		GroupBy   string `form:"group_by"`
	}
	CurvePoint struct {
		Time  int64   `json:"time"`
		Value float64 `json:"value"`
	}
	GrowthCurveResp struct {
		Metric string       `json:"metric"`
		Unit   string       `json:"unit"`
		Items  []CurvePoint `json:"items"`
	}
)

type (
	SleepListAtUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	SleepListAtReq struct {
		Date string `form:"date" binding:"required"`
	}
	SleepItem struct {
		SessionID  string `json:"session_id"`
		StartedAt  int64  `json:"started_at"`
		EndedAt    int64  `json:"ended_at"`
		DurationMs int64  `json:"duration_ms"`
	}
	SleepListAtResp struct {
		Items []SleepItem `json:"items"`
	}
)

// 睡眠计时（baby/daily）
type (
	SleepStartUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	SleepStartResp struct {
		SessionID string `json:"session_id"`
		StartedAt int64  `json:"started_at"`
	}
)

type (
	SleepStopUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	SleepStopReq struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	SleepStopResp struct {
		SessionID  string `json:"session_id"`
		StartedAt  int64  `json:"started_at"`
		EndedAt    int64  `json:"ended_at"`
		DurationMs int64  `json:"duration_ms"`
	}
)

type (
	SleepActiveUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	SleepActiveResp struct {
		SessionID string `json:"session_id"`
		StartedAt int64  `json:"started_at"`
	}
)

// 换尿布记录（baby/daily）
type (
	EnumItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	DiaperCreateUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	DiaperCreateReq struct {
		DiaperType      string `json:"diaper_type" binding:"required"`
		ChangeTime      int64  `json:"change_time" binding:"required"`
		PeeColor        string `json:"pee_color"`
		PoopColor       string `json:"poop_color"`
		PoopConsistency string `json:"poop_consistency"`
		Remark          string `json:"remark"`
	}
	DiaperCreateResp struct {
		DiaperID string `json:"diaper_id"`
		Message  string `json:"message"`
	}
)
type (
	DiaperUpdateUri struct {
		BabyID   string `uri:"baby_id" binding:"required"`
		DiaperID string `uri:"diaper_id" binding:"required"`
	}
	DiaperUpdateReq struct {
		DiaperType      string `json:"diaper_type" binding:"required"`
		ChangeTime      int64  `json:"change_time" binding:"required"`
		PeeColor        string `json:"pee_color"`
		PoopColor       string `json:"poop_color"`
		PoopConsistency string `json:"poop_consistency"`
		Remark          string `json:"remark"`
	}
	DiaperUpdateResp struct {
		Message string `json:"message"`
	}
)
type (
	DiaperGetAtUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	DiaperGetAtReq struct {
		Date string `form:"date" binding:"required"`
	}
	DiaperRecordResp struct {
		DiaperID        string    `json:"diaper_id"`
		ChangeTime      int64     `json:"change_time"`
		DiaperType      EnumItem  `json:"diaper_type"`
		PeeColor        *EnumItem `json:"pee_color"`
		PoopColor       *EnumItem `json:"poop_color"`
		PoopConsistency *EnumItem `json:"poop_consistency"`
		Remark          string    `json:"remark"`
	}
)

type (
	DiaperListAtUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	DiaperListAtReq struct {
		Date string `form:"date" binding:"required"`
	}
	DiaperItem struct {
		DiaperID        string    `json:"diaper_id"`
		ChangeTime      int64     `json:"change_time"`
		DiaperType      EnumItem  `json:"diaper_type"`
		PeeColor        *EnumItem `json:"pee_color"`
		PoopColor       *EnumItem `json:"poop_color"`
		PoopConsistency *EnumItem `json:"poop_consistency"`
		Remark          string    `json:"remark"`
	}
	DiaperListAtResp struct {
		Items []DiaperItem `json:"items"`
	}
)
type (
	DailyStatsUri struct {
		BabyID string `uri:"baby_id" binding:"required"`
	}
	DailyStatsReq struct {
		Date string `form:"date" binding:"required"`
	}
	DailyStatsResp struct {
		FeedingCount    int64 `json:"feeding_count"`
		SleepDurationMs int64 `json:"sleep_duration_ms"`
		DiaperCount     int64 `json:"diaper_count"`
	}
)
