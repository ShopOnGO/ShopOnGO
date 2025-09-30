package question

type addQuestionRequest struct {
	ProductID uint   `json:"product_id"`
	QuestionText     string `json:"question_text"`
}

type answerQuestionRequest struct {
	AnswerText string `json:"answer_text"`
}