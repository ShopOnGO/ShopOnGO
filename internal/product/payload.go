package product

import (
	"github.com/ShopOnGO/ShopOnGO/internal/productVariant"
)

type addProductRequest struct {
	Name        string 	`json:"name"`
	Description string 	`json:"description"`
	Material    string 	`gorm:"type:varchar(200)"`
	IsActive    bool   	`json:"is_active"`

	CategoryID 	uint   	`json:"category_id"`
	BrandID    	uint   	`json:"brand_id"`

	ImageKeys  []string `json:"image_keys"`
	VideoKeys  []string `json:"video_keys"`

	Variants []productVariant.AddProductVariantRequest `json:"variants"`
}

type productCreatedEvent struct {
	Action  string         	   `json:"action"`
	Product addProductRequest  `json:"product"`
}
