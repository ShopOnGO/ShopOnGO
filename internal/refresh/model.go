package refresh

import "time"

type RefreshTokenRecord struct {
	Token     string    `gorm:"primaryKey"`
	Email     string    `gorm:"index;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
}
