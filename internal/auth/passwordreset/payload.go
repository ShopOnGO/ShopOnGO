package passwordreset

type ResetRequest struct {
    Email    string `json:"email" validate:"required,email"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
    Code  string `json:"code" validate:"required"`
}

type ResetPasswordRequest struct {
	Email    		string `json:"email" validate:"required,email"`
    Code        	string `json:"code" validate:"required"`
	NewPassword 	string `json:"new_password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
