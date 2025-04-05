package comment

type CommentRepository interface {
	GetCommentsByReviewID(reviewID uint) ([]Comment, error)
	AddComment(comment *Comment) error
	UpdateComment(comment *Comment) error
	DeleteComment(comment *Comment) error
	IncrementLikesCount(commentID uint) error
	GetReplies(commentID uint) ([]Comment, error)
}