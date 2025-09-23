package productVariant

import "github.com/shopspring/decimal"

type AddProductVariantRequest struct {
	SKU           string          	`json:"sku"`
	Price         decimal.Decimal 	`json:"price"`
	Discount      decimal.Decimal 	`json:"discount"`
	Sizes         string        	`json:"sizes"`
	Colors        string        	`json:"colors"`
	Stock         uint32          	`json:"stock"`
	Barcode       string          	`json:"barcode"`
	IsActive      bool            	`json:"is_active"`
	Images        []string        	`json:"images"`
	MinOrder      uint            	`json:"min_order"`
	Dimensions    string          	`json:"dimensions"`
}


type productVariantCreatedEvent struct {
	Action  		string                   `json:"action"`
	ProductID		uint					 `json:"product_id"`
	ProductVariant 	AddProductVariantRequest `json:"product_variant"`
	UserID 			uint                     `json:"user_id"`
}
