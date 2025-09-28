package review

import (
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/internal/user"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID          uint    `gorm:"not null;uniqueIndex:idx_user_product" json:"user_id"`
	ProductID   	uint    `gorm:"not null;uniqueIndex:idx_user_product" json:"product_id"`
	Rating          int16	`gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	LikesCount		int    	`gorm:"default:0" json:"likes_count"`
	Comment         string	`gorm:"not null" json:"comment"`

	User            user.User      					`gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	ProductVariant  productVariant.ProductVariant	`gorm:"foreignKey:ProductVariantID;constraint:OnDelete:CASCADE" json:"product_variant"`
}
