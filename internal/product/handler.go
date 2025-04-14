package product

import (
	"encoding/json"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	"github.com/gorilla/mux"
)

type ProductHandlerDeps struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

type ProductHandler struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

func NewProductHandler(router *mux.Router, deps ProductHandlerDeps) {
	handler := &ProductHandler{
		Config: deps.Config,
		Kafka:  deps.Kafka,
	}
	router.HandleFunc("/products", handler.AddProduct()).Methods("POST")
}

func (h *ProductHandler) AddProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Price <= 0 || req.CategoryID == 0 || req.BrandID == 0 {
			http.Error(w, "missing or invalid required fields", http.StatusBadRequest)
			return
		}

		event := productCreatedEvent{
			Action: "create",
			Product: req,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "failed to marshal event", http.StatusInternalServerError)
			return
		}

		if err := h.Kafka.Produce(r.Context(), []byte("product-create"), eventBytes); err != nil {
			logger.Errorf("Kafka produce error: %v", err)
			http.Error(w, "failed to send Kafka message", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]interface{}{
			"status":  "product created and event sent",
			"product": req,
		}, http.StatusCreated)
	}
}

