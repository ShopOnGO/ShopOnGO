package user

import "gorm.io/gorm"

type User struct {
	gorm.Model   `swaggerignore:"true"` // Включает ID, CreatedAt, UpdatedAt, DeletedAt
	Name         string `gorm:"not null"`
	Email        string `gorm:"unique;not null;index"`
	PasswordHash string `gorm:"default:null"`
	Role         string `gorm:"not null;default:'buyer'"` // "admin", "seller", "buyer"
	Provider     string `gorm:"not null;default:'local'"` // "local", "google", "facebook"
	Status       string `gorm:"not null;default:'active'"` // "active", "banned", "deleted"
	Phone        string `gorm:"default:null"`
	ProfileImage string `gorm:"default:null"`

	StoreName    *string `gorm:"default:null"`
	StoreAddress *string `gorm:"default:null"`
	StorePhone   *string `gorm:"default:null"`

	AcceptTerms  bool `gorm:"not null;default:false"`
}
