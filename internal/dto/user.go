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
		PartnerID string `json:"partner_id"`
	}
	PartnerGetResp struct {
		PartnerID string `json:"partner_id"`
	}
)
