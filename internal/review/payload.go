package review

type addReviewRequest struct {
	ProductVariantID uint   `json:"product_variant_id"`
	Rating           int16  `json:"rating"`
	Comment          string `json:"comment"`
}

