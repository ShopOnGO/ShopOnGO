package passwordreset

type RequestResetRequest struct {
    Email string `json:"email"`
}

type VerifyCodeRequest struct {
    Email string `json:"email"`
    Code  string `json:"code"`
}
