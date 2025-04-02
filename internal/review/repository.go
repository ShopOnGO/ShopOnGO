package review

import (
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
)

type ReviewRepository struct {
	Db *db.Db
}

func NewRepository(db *db.Db) *ReviewRepository {
	return &ReviewRepository{
		Db: db,
	}
}

func (r *ReviewRepository) CreateReview(review *Review) error {
	return r.Db.Create(review).Error
}

func (r *ReviewRepository) GetReviewByID(id uint) (*Review, error) {
	var review Review
	err := r.Db.First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetReviewsByProductVariantID(productVariantID uint) ([]Review, error) {
	var reviews []Review
	err := r.Db.Where("product_variant_id = ?", productVariantID).Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewRepository) UpdateReview(review *Review) error {
	return r.Db.Save(review).Error
}

func (r *ReviewRepository) DeleteReview(review *Review) error {
	return r.Db.Delete(review).Error
}
