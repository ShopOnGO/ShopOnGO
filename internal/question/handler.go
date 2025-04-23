package question

import (
	"encoding/json"
	"fmt"
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

	router.Handle("/questions", middleware.AuthOrGuest(handler.AddQuestion(), deps.Config)).Methods("POST")
	router.Handle("/questions/{id}", middleware.IsAuthed(handler.AnswerQuestion(), deps.Config)).Methods("PUT")
	router.Handle("/questions/{id}", middleware.IsAuthed(handler.DeleteQuestion(), deps.Config)).Methods("DELETE")
	router.Handle("/questions/{id}/likes", middleware.IsAuthed(handler.AddLikeToQuestion(), deps.Config)).Methods("PUT")
	router.Handle("/questions/{id}/unlikes", middleware.IsAuthed(handler.RemoveLikeToQuestion(), deps.Config)).Methods("PUT")
}

func (qh *QuestionHandler) AddQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, guestID, err := getUserOrGuestID(r)
		// logger.Infof("user=%v, guest=%v", UserID, GuestID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}
		var req addQuestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		
		// Формирование события
		author := map[string]interface{}{}
		if userID != nil {
			author["user_id"] = *userID
		} else {
			author["guest_id"] = fmt.Sprintf("%x", guestID)
		}

		event := map[string]interface{}{
			"action":             "created",
			"product_variant_id": req.ProductVariantID,
			"question_text":      req.QuestionText,
			"author":             author,
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
		idStr := strings.TrimPrefix(r.URL.Path, "/questions/")
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
		idStr := strings.TrimPrefix(r.URL.Path, "/questions/")
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


func (qh *QuestionHandler) AddLikeToQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(middleware.ContextUserIDKey)
		userID, ok := userIDVal.(uint)
		if !ok || userID == 0 {
			http.Error(w, "invalid or missing user_id", http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]
		questionID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || questionID == 0 {
			http.Error(w, "invalid question id", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":      "addLike",
			"question_id": questionID,
			"user_id":     userID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing like event", http.StatusInternalServerError)
			return
		}

		key := []byte("question")
		if err := qh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka like event: %v", err)
			http.Error(w, "failed to send like event to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "question like event sent"}, http.StatusOK)
	}
}

func (qh *QuestionHandler) RemoveLikeToQuestion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(middleware.ContextUserIDKey)
		userID, ok := userIDVal.(uint)
		if !ok || userID == 0 {
			http.Error(w, "invalid or missing user_id", http.StatusForbidden)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]
		questionID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || questionID == 0 {
			http.Error(w, "invalid question id", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":      "removeLike",
			"question_id": questionID,
			"user_id":     userID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error marshalling event", http.StatusInternalServerError)
			return
		}

		key := []byte("question")
		if err := qh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka removelike event: %v", err)
			http.Error(w, "failed to send removelike event to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "question removelike event sent"}, http.StatusOK)
	}
}


func getUserOrGuestID(r *http.Request) (*uint, []byte, error) {
	userIDVal := r.Context().Value(middleware.ContextUserIDKey)
	var userID *uint
	if id, ok := userIDVal.(uint); ok && id != 0 {
		userID = &id
	}

	guestIDVal := r.Context().Value(middleware.ContextGuestIDKey)
	var guestID []byte
	if id, ok := guestIDVal.([]byte); ok {
		guestID = id
	}

	if userID == nil && len(guestID) == 0 {
		return nil, nil, fmt.Errorf("не удалось определить пользователя: no user or guest ID in context")
	}

	logger.Infof("Raw guestID: %v", guestID)

	return userID, guestID, nil
}