package review

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"

	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID             uint      `gorm:"not null" json:"user_id"`
	User               user.User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	ProductVariantID   uint      `gorm:"not null" json:"product_variant_id"`
	Rating             int16     `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`

	ProductVariant     productVariant.ProductVariant `gorm:"foreignKey:ProductVariantID;constraint:OnDelete:CASCADE" json:"product_variant"`
}
