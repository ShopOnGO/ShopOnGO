package user

import "gorm.io/gorm"

type User struct {
	gorm.Model `swaggerignore:"true"`
	Email      string `gorm:"index"`
	Password   string
	Name       string
	Role       string `gorm:"default:'buyer'"` // "admin", "seller", "buyer"
}
