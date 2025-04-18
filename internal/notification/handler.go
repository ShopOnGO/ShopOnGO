package notification

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	"github.com/gorilla/mux"
)

type NotificationHandlerDeps struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

type NotificationHandler struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

func NewNotificationHandler(router *mux.Router, deps NotificationHandlerDeps) {
	handler := &NotificationHandler{
		Config: deps.Config,
		Kafka:  deps.Kafka,
	}
	router.Handle("/notifications", middleware.IsAuthed(handler.AddNotification(), deps.Config)).Methods("POST")
	//router.Handle("/notifications/{id}", middleware.IsAuthed(handler.UpdateReview(), deps.Config)).Methods("PUT")
	//router.Handle("/notifications/{id}", middleware.IsAuthed(handler.DeleteReview(), deps.Config)).Methods("DELETE")
}

func (h *NotificationHandler) AddNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddNotification
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		userID := r.Context().Value(middleware.ContextUserIDKey)

		if req.Category == "" || userID == 0 {
			http.Error(w, "category and user_id are required", http.StatusBadRequest)
			return
		}
		event := map[string]interface{}{
			"action":   "create",
			"category": req.Category,
			"subtype":  req.Subtype,
			"userID":   userID,
			"wasInDlq": false,
			"payload":  req.Payload,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "error processing event", http.StatusInternalServerError)
			return
		}

		key := []byte(fmt.Sprintf("notification-AddNote:%v", userID))
		if err := h.Kafka.Produce(r.Context(), key, eventBytes); err != nil {
			logger.Errorf("Error producing Kafka message: %v", err)
			http.Error(w, "failed to send message to kafka", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]string{"status": "review creation event sent"}, http.StatusOK)
	}
}
