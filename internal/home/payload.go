package home

import (
	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
)

type HomeData struct {
	Categories       []category.Category `json:"categories"`
	Brands           []brand.Brand       `json:"featured_brands"`
	// Promotions       []Promotion `json:"promotions"`
}
