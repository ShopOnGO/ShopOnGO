package passwordreset

type ResetRequest struct {
    Email string `json:"email"`
}

type VerifyCodeRequest struct {
    Email string `json:"email"`
    Code  string `json:"code"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
    Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}
