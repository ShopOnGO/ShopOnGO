package home

import (
	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
	"github.com/ShopOnGO/ShopOnGO/internal/product"
)

type HomeData struct {
	Categories       []category.Category `json:"categories"`
	FeaturedProducts []product.Product   `json:"featured_products"`
	Brands           []brand.Brand       `json:"featured_brands"`
	// Promotions       []Promotion `json:"promotions"`
}
