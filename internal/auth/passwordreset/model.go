package passwordreset

import "time"

type ResetToken struct {
    Email     string    `json:"email"`
    Code      string    `json:"code"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
}