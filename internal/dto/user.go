package dto

type (
	LoginReq struct {
		Account   string `json:"account"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		Code      string `json:"code"`
		LoginType string `json:"login_type"`
	}
	LoginResp struct {
		Token string `json:"token"`
	}
)

type (
	GetCodeReq struct {
		Email string `json:"email"`
	}
	GetCodeResp struct {
		Code string `json:"code"`
	}
)

type (
	RegisterReq struct {
		Account  string `json:"account"`
		Password string `json:"password"`
		Username string `json:"username"`
		Gender   string `json:"gender"`
		Email    string `json:"email"`
		Code     string `json:"code"`
	}
	RegisterResp struct {
		Message string `json:"message"`
	}
)

type (
	ResetPasswordReq struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	ResetPasswordResp struct {
		Message string `json:"message"`
	}
)

type (
	UpdateUserAdditionReq struct {
		Phone      *string `json:"phone"`
		Occupation *string `json:"occupation"`
		Gender     *string `json:"gender"`
		Province   *string `json:"province"`
		City       *string `json:"city"`
		Birthday   *string `json:"birthday"`
	}
	UpdateUserAdditionResp struct {
		Message string `json:"message"`
	}
)

type (
	UpdateAvatarReq struct {
		Avatar string `json:"avatar"`
	}
	UpdateAvatarResp struct {
		Message string `json:"message"`
	}
)

type (
	PartnerBindReq struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}
	PartnerBindResp struct {
		PartnerID       string `json:"partner_id"`
		PartnerUsername string `json:"partner_username"`
		PartnerAvatar   string `json:"partner_avatar"`
	}
	PartnerGetResp struct {
		PartnerID       string `json:"partner_id"`
		PartnerUsername string `json:"partner_username"`
		PartnerAvatar   string `json:"partner_avatar"`
	}
	MyProfileResp struct {
		UserID     string `json:"user_id"`
		Account    string `json:"account"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		Gender     string `json:"gender"`
		Avatar     string `json:"avatar"`
		Phone      string `json:"phone"`
		Occupation string `json:"occupation"`
		Birthday   int64  `json:"birthday"`
		Province   string `json:"province"`
		City       string `json:"city"`
		Ctime      int64  `json:"ctime"`
		Utime      int64  `json:"utime"`
		PartnerID  string `json:"partner_id"`
	}
)

type (
	FollowReq struct {
		TargetUserID string `uri:"target_user_id"`
	}
	FollowResp struct {
		Message string `json:"message"`
	}
	FollowingListReq struct {
		UserID   string `form:"user_id"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	FollowingUserItem struct {
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		Avatar     string `json:"avatar"`
		FollowTime int64  `json:"follow_time"`
	}
	FollowingListResp struct {
		List    []FollowingUserItem `json:"list"`
		HasMore bool                `json:"has_more"`
	}
)

type (
	FollowersListReq struct {
		UserID   string `form:"user_id"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	FollowersListResp struct {
		List    []FollowingUserItem `json:"list"`
		HasMore bool                `json:"has_more"`
	}
)

type (
	AdminListUsersReq struct {
		Keyword  string `form:"keyword"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	AdminUserItem struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	AdminListUsersResp struct {
		List    []AdminUserItem `json:"list"`
		HasMore bool            `json:"has_more"`
	}
	AdminPromoteUri struct {
		UserID string `uri:"user_id"`
	}
)
