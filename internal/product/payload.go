package product

type addProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Discount    int64  `json:"discount"`
	IsActive    bool   `json:"is_active"`

	CategoryID 	uint   `json:"category_id"`
	BrandID    	uint   `json:"brand_id"`

	Images   	string `json:"images"`
	VideoURL 	string `json:"video_url"`
}

type productCreatedEvent struct {
	Action  string         	   `json:"action"`
	Product addProductRequest  `json:"product"`
}
