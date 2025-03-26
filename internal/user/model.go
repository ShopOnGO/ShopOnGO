package user

import "gorm.io/gorm"

type User struct {
	gorm.Model `swaggerignore:"true"`
	Email      string `gorm:"index"`
	Password   string `gorm:"default:''"`
	Name       string
	Role       string `gorm:"default:'buyer'"` // "admin", "seller", "buyer"
	Provider   string `gorm:"default:'local'"` // "google" или "local"
}
