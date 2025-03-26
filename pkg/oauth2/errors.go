package oauth2

import "errors"

var (
	ErrInvalidOrExpiredRefreshToken = errors.New("invalid or expired refresh token")
	ErrFailedToCreateNewTokens = errors.New("failed to create new tokens")
	ErrFailedToStoreRefreshToken = errors.New("failed to store refresh token")
	ErrRefreshTokenCookieNotFound = errors.New("refresh token cookie not found")
)
