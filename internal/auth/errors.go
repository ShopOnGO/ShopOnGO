package auth

const (
	ErrUserExists       = "user exists"
	ErrWrongCredentials = "wrong email or password"
	ErrRefreshTokenNotFound = "refresh token not found"
	ErrFailedToExchangeToken = "failed to exchange token"
	ErrFailedToGetUserInfo = "failed to get user info"
	ErrFailedToGetUserRole = "failed to get user role"
	ErrFailedToUpdateUserRole = "failed to update user role"
	ErrFailedToDecodeUserInfo = "failed to decode user info"
	ErrFailedToGenerateTokens = "failed to generate tokens"
	ErrFailedRefreshTokenNotFound = "failed refresh token not found"
	ErrFailedToLogout = "Failed to logout"
	ErrUserNotFound = "user not found"
	ErrFailedToFindUser = "failed to find user"
	ErrInvalidRequestData = "invalid request data"
	FailedToHashNewPassword = "failed to hash new password"
	FailedToUpdatePassword = "failed to update password"
	ErrRecordNotFound = "record not found"
	ErrorCreatingorFindingUser = "error creating or finding user"
)
