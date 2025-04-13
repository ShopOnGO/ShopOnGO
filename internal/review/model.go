package review

import (
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/internal/user"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID             uint      `gorm:"not null;uniqueIndex:idx_user_product" json:"user_id"`
	ProductVariantID   uint      `gorm:"not null;uniqueIndex:idx_user_product" json:"product_variant_id"`
	Rating             int16     `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`

	User               user.User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	ProductVariant     productVariant.ProductVariant `gorm:"foreignKey:ProductVariantID;constraint:OnDelete:CASCADE" json:"product_variant"`
}
