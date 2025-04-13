package question

import (
	"gorm.io/gorm"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
)

type Question struct {
	gorm.Model
	UserID           uint      `gorm:"not null" json:"user_id"`
	User             user.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
	ProductVariantID uint      `gorm:"not null" json:"product_variant_id"`
	QuestionText     string    `gorm:"not null" json:"question_text"`
	AnswerText       string    `json:"answer_text"`
}
