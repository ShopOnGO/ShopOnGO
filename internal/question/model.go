package question

import (
	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"gorm.io/gorm"
)

type Question struct {
	gorm.Model
	UserID          uint      `gorm:"index" json:"user_id"`
	GuestID   		[]byte    `gorm:"type:bytea;index"`
	ProductID		uint      `gorm:"not null" json:"product_id"`
	QuestionText    string    `gorm:"not null" json:"question_text"`
	AnswerText      string    `json:"answer_text"`
	LikesCount		int       `gorm:"default:0" json:"likes_count"`

	User            user.User 		`gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
}
