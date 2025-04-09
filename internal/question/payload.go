package question

type addQuestionRequest struct {
	ProductVariantID uint   `json:"product_variant_id"`
	QuestionText     string `json:"question_text"`
}

type answerQuestionRequest struct {
	AnswerText string `json:"answer_text"`
}