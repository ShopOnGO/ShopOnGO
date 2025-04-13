package comment

import (
	"time"

	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
	"gorm.io/gorm"
)

type Comment struct {
    gorm.Model
    // ID родительского комментария, если это ответ
    ParentCommentID *uint      `gorm:"index" json:"parent_comment_id,omitempty"` 
    ReviewID        uint       `gorm:"not null" json:"review_id"`
    UserID          uint       `gorm:"not null" json:"user_id"`
    Comment         string     `gorm:"type:text;not null" json:"comment"`
    PublicationDate time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"publication_date"`
    IsApproved      bool       `gorm:"not null;default:false" json:"is_approved"`
    LikesCount      int        `gorm:"default:0" json:"likes_count"`

    Review          review.Review     `gorm:"foreignKey:ReviewID" json:"review"`
    User           	user.User       `gorm:"foreignKey:UserID" json:"user"`
    ParentComment   *Comment   `gorm:"foreignKey:ParentCommentID" json:"parent_comment,omitempty"`
    Replies         []Comment  `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
}
