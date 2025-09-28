package review

type addReviewRequest struct {
	ProductID 	uint   `json:"product_id"`
	Rating      int16  `json:"rating"`
	Comment     string `json:"comment"`
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