package favorites

import (

	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
	"gorm.io/gorm"
)

type Favorite struct {
	gorm.Model
	UserID           uint `gorm:"not null;index" json:"user_id"`
	ProductVariantID uint `gorm:"not null;index" json:"product_id"`

	User           user.User                     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	ProductVariant productVariant.ProductVariant `gorm:"foreignKey:ProductVariantID;constraint:OnDelete:CASCADE" json:"product_variant"`
}
