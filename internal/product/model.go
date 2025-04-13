package product

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/productVariant"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/review"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model

	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Price       int64  `gorm:"not null" json:"price"`
	Discount    int64  `gorm:"default:0" json:"discount"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// 🔹 Внешние ключи
	CategoryID uint              `gorm:"not null" json:"category_id"`
	Category   category.Category `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE"`

	BrandID uint        `gorm:"not null" json:"brand_id"`
	Brand   brand.Brand `gorm:"foreignKey:BrandID;constraint:OnDelete:CASCADE"`

	Variants []productVariant.ProductVariant `gorm:"foreignKey:ProductID"` // Ссылка на варианты продукта

	// 🔹 Дополнительные данные
	Images   string `gorm:"type:json" json:"images"`            // Храним ссылки на изображения JSON-массивом
	VideoURL string `gorm:"type:varchar(255)" json:"video_url"` // Видеообзор

	Reviews []review.Review `gorm:"-" json:"reviews"`
}

//todo category_id (3)
