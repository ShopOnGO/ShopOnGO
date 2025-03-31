package home

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
)

type HomeData struct {
	Categories       []category.Category `json:"categories"`
	FeaturedProducts []product.Product   `json:"featured_products"`
	Brands           []brand.Brand       `json:"featured_brands"`
	// Promotions       []Promotion `json:"promotions"`
}
