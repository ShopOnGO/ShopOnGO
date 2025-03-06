package home

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
)

type HomeData struct {
	Categories       []category.Category `json:"categories"`
	FeaturedProducts []product.Product   `json:"featured_products"`
	// Promotions       []Promotion `json:"promotions"`
}
