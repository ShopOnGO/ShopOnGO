package review

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"github.com/gorilla/mux"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/res"
)

type ReviewHandlerDeps struct {
	Config *configs.Config
	Kafka *kafkaService.KafkaService
}

type ReviewHandler struct {
	Config *configs.Config
	Kafka *kafkaService.KafkaService
}

func NewReviewHandler(router *mux.Router, deps ReviewHandlerDeps){
	handler := &ReviewHandler{
		Config:     deps.Config,
		Kafka: 		deps.Kafka,
	}
	router.Handle("/review", middleware.IsAuthed(handler.AddReview(), deps.Config)).Methods("POST")
	router.Handle("/review/{id}", middleware.IsAuthed(handler.UpdateReview(), deps.Config)).Methods("PUT")
	router.Handle("/review/{id}", middleware.IsAuthed(handler.DeleteReview(), deps.Config)).Methods("DELETE")
}


func (rh *ReviewHandler) AddReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		userID := r.Context().Value(middleware.ContextUserIDKey)
		if req.ProductVariantID == 0 || userID == 0 {
			http.Error(w, "product_variant_id and user_id are required", http.StatusBadRequest)
			return
		}
		
		event := map[string]interface{}{
			"action":             "created",
			"product_variant_id": req.ProductVariantID,
			"user_id":            userID,
			"rating":             req.Rating,
			"comment":            req.Comment,
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		key := []byte("review-AddReview")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "review creation event sent"}, http.StatusOK)
	}
}

func (rh *ReviewHandler) UpdateReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлечение идентификатора отзыва из URL ("/reviews/{id}")
		idStr := strings.TrimPrefix(r.URL.Path, "/review/")
		reviewID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || reviewID == 0 {
			http.Error(w, "invalid review id", http.StatusBadRequest)
			return
		}
		var req updateReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		event := map[string]interface{}{
			"action":    "updated",
			"review_id": reviewID,
		}
		// Добавляем только переданные поля
		if req.Rating != 0 {
			event["rating"] = req.Rating
		}
		if req.Comment != "" {
			event["comment"] = req.Comment
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}
		key := []byte("review-id-" + strconv.FormatUint(reviewID, 10))
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}
		res.Json(w, map[string]string{"status": "review update event sent"}, http.StatusOK)
	}
}

func (rh *ReviewHandler) DeleteReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/review/")
		reviewID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || reviewID == 0 {
			http.Error(w, "invalid review id", http.StatusBadRequest)
			return
		}
		event := map[string]interface{}{
			"action":    "deleted",
			"review_id": reviewID,
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}
		key := []byte("review-id-" + strconv.FormatUint(reviewID, 10))
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}
		res.Json(w, map[string]string{"status": "review deletion event sent"}, http.StatusOK)
	}
}
