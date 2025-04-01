package cart

import (
	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model   `swaggerignore:"true"`
	UserID    uint       `gorm:"index"`
	GuestID   []byte 	 `gorm:"type:bytea;index"`
	CartItems []CartItem `gorm:"foreignKey:CartID"`
}

 type CartItem struct {
	gorm.Model   `swaggerignore:"true"`
	CartID           uint `gorm:"not null;index"`
	ProductVariantID uint `gorm:"not null;index"`
	Quantity         int `gorm:"not null;default:1"`
}