package cart

type AddCartItemRequest struct {
	ProductVariantID uint `json:"product_variant_id" binding:"required"`
	Quantity         int  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	ProductVariantID uint `json:"product_variant_id" binding:"required"`
	Quantity         int  `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	UserID   uint             `json:"user_id,omitempty"`
	GuestID  []byte            `json:"guest_id,omitempty"`
	Items    []CartItemResponse `json:"items"`
}

type CartItemResponse struct {
	ProductVariantID uint `json:"product_variant_id"`
	Quantity         int  `json:"quantity"`
}

type RemoveCartItemRequest struct {
	ProductVariantID uint `json:"product_variant_id" binding:"required"`
}
