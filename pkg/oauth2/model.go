package oauth2


type RefreshTokenData struct {
	UserID uint `json:"user_id"`
	Role   string `json:"role"`
}