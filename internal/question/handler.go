package question

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	"github.com/gorilla/mux"
)

type QuestionHandlerDeps struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

type QuestionHandler struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

// NewQuestionHandler регистрирует пути для работы с вопросами.
func NewQuestionHandler(router *mux.Router, deps QuestionHandlerDeps) {
	handler := &QuestionHandler{
		Config: deps.Config,
		Kafka:  deps.Kafka,
	}

	router.Handle("/questions", handler.AddQuestion()).Methods("POST")
	router.Handle("/questions/{id}", middleware.IsAuthed(handler.AnswerQuestion(), deps.Config)).Methods("PUT")
	router.Handle("/questions/{id}", middleware.IsAuthed(handler.DeleteQuestion(), deps.Config)).Methods("DELETE")	
}

func (qh *QuestionHandler) AddQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addQuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		
		// Формирование события
		event := map[string]interface{}{
			"action":             "created",
			"product_variant_id": req.ProductVariantID,
			"question_text":      req.QuestionText,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		// Ключ сообщения можно задать, например, так:
		key := []byte("question")
		if err := qh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "question creation event sent"}, http.StatusOK)
	}
}

func (qh *QuestionHandler) AnswerQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ожидается, что URL имеет вид /question/{id}
		idStr := strings.TrimPrefix(r.URL.Path, "/question/")
		questionID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || questionID == 0 {
			http.Error(w, "invalid question id", http.StatusBadRequest)
			return
		}

		var req answerQuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.AnswerText == "" {
			http.Error(w, "answer_text is required", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":      "answered",
			"question_id": questionID,
			"answer_text": req.AnswerText,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		key := []byte("question")
		if err := qh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "question answer event sent"}, http.StatusOK)
	}
}

func (qh *QuestionHandler) DeleteQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/question/")
		questionID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || questionID == 0 {
			http.Error(w, "invalid question id", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":      "deleted",
			"question_id": questionID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		key := []byte("question")
		if err := qh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "question deletion event sent"}, http.StatusOK)
	}
}
