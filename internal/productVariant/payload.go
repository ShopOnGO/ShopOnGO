package productVariant

import "github.com/shopspring/decimal"

type addProductVariantRequest struct {
	SKU           string          `json:"sku"`
	Price         decimal.Decimal `json:"price"`
	Discount      decimal.Decimal `json:"discount"`
	Sizes         []uint32        `json:"sizes"`
	Colors        []string        `json:"colors"`
	Stock         uint32          `json:"stock"`
	Material      string          `json:"material"`
	Barcode       string          `json:"barcode"`
	IsActive      bool            `json:"is_active"`
	Images        []string        `json:"images"`
	MinOrder      uint            `json:"min_order"`
	Dimensions    string          `json:"dimensions"`
}


type productVariantCreatedEvent struct {
	Action  		string                   `json:"action"`
	ProductID		uint					 `json:"product_id"`
	ProductVariant 	addProductVariantRequest `json:"product_variant"`
	UserID 			uint                     `json:"user_id"`
}
