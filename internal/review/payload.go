package review

type addReviewRequest struct {
	ProductVariantID uint   `json:"product_variant_id"`
	Rating           int16  `json:"rating"`
	LikesCount		 int    `gorm:"default:0" json:"likes_count"`
	Comment          string `json:"comment"`
}

type reviewCreatedEvent struct {
	Action  string         	   `json:"action"`
	Review  addReviewRequest   `json:"product"`
	UserID  uint			   `json:"user_id"`
}

type updateReviewRequest struct {
	Rating  int16  `json:"rating"`
	Comment string `json:"comment"`
}