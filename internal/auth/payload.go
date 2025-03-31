package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type RegisterRequest struct {
	LoginRequest
	Name string `json:"name" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	OldPassword 	string `json:"old_password" validate:"required"`
	NewPassword 	string `json:"new_password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type ChangeRoleRequest struct {
    Email           string `json:"email" validate:"required,email"`
    CurrentPassword string `json:"current_password" validate:"required"`
    Phone      		string `json:"phone,omitempty" validate:"omitempty,e164"`
	NewRole         string `json:"new_role" validate:"required,oneof=buyer seller moderator"`
    
    // Поля для продавца
    StoreName       string `json:"store_name" validate:"required_if=NewRole seller"`
    StoreAddress    string `json:"store_address" validate:"required_if=NewRole seller"`
    StorePhone      string `json:"store_phone,omitempty" validate:"omitempty,e164"`
    
    // Согласие с условиями
    AcceptTerms     bool   `json:"accept_terms,omitempty" validate:"omitempty,eq=true"`
}

