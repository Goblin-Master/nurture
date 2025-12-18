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
		Token    string `json:"token"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
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
