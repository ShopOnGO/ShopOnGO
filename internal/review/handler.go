package review

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
)

type ReviewHandlerDeps struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

type ReviewHandler struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

func NewReviewHandler(router *mux.Router, deps ReviewHandlerDeps) {
	handler := &ReviewHandler{
		Config: deps.Config,
		Kafka:  deps.Kafka,
	}
	router.Handle("/reviews", middleware.IsAuthed(handler.AddReview(), deps.Config)).Methods("POST")
	router.Handle("/reviews/{id}", middleware.IsAuthed(handler.UpdateReview(), deps.Config)).Methods("PUT")
	router.Handle("/reviews/{id}", middleware.IsAuthed(handler.DeleteReview(), deps.Config)).Methods("DELETE")
	router.Handle("/reviews/{id}/likes", middleware.IsAuthed(handler.AddLikeToReview(), deps.Config)).Methods("PUT")
	router.Handle("/reviews/{id}/unlikes", middleware.IsAuthed(handler.RemoveLikeToReview(), deps.Config)).Methods("PUT")
}

// AddReview добавляет новый отзыв
// @Summary Добавить отзыв
// @Description Создание нового отзыва о товаре
// @Tags Reviews
// @Accept json
// @Produce json
// @Param review body addReviewRequest true "Данные для создания отзыва"
// @Success 200 {object} map[string]string "status: review creation event sent"
// @Failure 400 {string} string "invalid request body / invalid user_id / product_variant_id and user_id are required"
// @Failure 500 {string} string "error processing event / failed to send message to kafka"
// @Security BearerAuth
// @Router /reviews [post]
func (rh *ReviewHandler) AddReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		if req.ProductVariantID == 0 || userID == 0 {
			http.Error(w, "product_variant_id and user_id are required", http.StatusBadRequest)
			return
		}

		event := reviewCreatedEvent{
			Action: "create",
			Review: req,
			UserID: userID,
		}
		logger.Info(event)

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		key := []byte("review")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "review creation event sent"}, http.StatusOK)
	}
}

// UpdateReview обновляет существующий отзыв
// @Summary Обновить отзыв
// @Description Обновление рейтинга или комментария отзыва
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Param review body updateReviewRequest true "Данные для обновления отзыва"
// @Success 200 {object} map[string]string "status: review update event sent"
// @Failure 400 {string} string "invalid review id / invalid request body / invalid user_id"
// @Failure 500 {string} string "error processing event / failed to send message to kafka"
// @Security BearerAuth
// @Router /reviews/{id} [put]
func (rh *ReviewHandler) UpdateReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		if userID == 0 {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/reviews/")
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
			"action":    "update",
			"review_id": reviewID,
			"user_id":   userID,
		}

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

		key := []byte("review")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}
		res.Json(w, map[string]string{"status": "review update event sent"}, http.StatusOK)
	}
}

// DeleteReview удаляет отзыв
// @Summary Удалить отзыв
// @Description Удаление отзыва по ID
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 200 {object} map[string]string "status: review deletion event sent"
// @Failure 400 {string} string "invalid review id"
// @Failure 500 {string} string "error processing event / failed to send message to kafka"
// @Security BearerAuth
// @Router /reviews/{id} [delete]
func (rh *ReviewHandler) DeleteReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/reviews/")
		reviewID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || reviewID == 0 {
			http.Error(w, "invalid review id", http.StatusBadRequest)
			return
		}
		event := map[string]interface{}{
			"action":    "delete",
			"review_id": reviewID,
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}
		key := []byte("review")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}
		res.Json(w, map[string]string{"status": "review deletion event sent"}, http.StatusOK)
	}
}

// AddLikeToReview добавляет лайк отзыву
// @Summary Поставить лайк отзыву
// @Description Добавление лайка к отзыву пользователем
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 200 {object} map[string]string "status: review like event sent"
// @Failure 400 {string} string "invalid review id / invalid or missing user_id"
// @Failure 500 {string} string "error marshalling event / failed to send like event to kafka"
// @Security BearerAuth
// @Router /reviews/{id}/likes [put]
func (rh *ReviewHandler) AddLikeToReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok || userID == 0 {
			http.Error(w, "invalid or missing user_id", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]
		reviewID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || reviewID == 0 {
			http.Error(w, "invalid review id", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":    "addLike",
			"review_id": reviewID,
			"user_id":   userID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error marshalling event", http.StatusInternalServerError)
			return
		}

		key := []byte("review")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka like event: %v", err)
			http.Error(w, "failed to send like event to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "review like event sent"}, http.StatusOK)
	}
}

// RemoveLikeToReview убирает лайк с отзыва
// @Summary Убрать лайк с отзыва
// @Description Удаление лайка отзыва пользователем
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID отзыва"
// @Success 200 {object} map[string]string "status: review removelike event sent"
// @Failure 400 {string} string "invalid review id / invalid or missing user_id"
// @Failure 500 {string} string "error marshalling event / failed to send removelike event to kafka"
// @Security BearerAuth
// @Router /reviews/{id}/unlikes [put]
func (rh *ReviewHandler) RemoveLikeToReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint)
		if !ok || userID == 0 {
			http.Error(w, "invalid or missing user_id", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["id"]
		reviewID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || reviewID == 0 {
			http.Error(w, "invalid review id", http.StatusBadRequest)
			return
		}

		event := map[string]interface{}{
			"action":    "removeLike",
			"review_id": reviewID,
			"user_id":   userID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error marshalling event", http.StatusInternalServerError)
			return
		}

		key := []byte("review")
		if err := rh.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka removelike event: %v", err)
			http.Error(w, "failed to send removelike event to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "review removelike event sent"}, http.StatusOK)
	}
}
