package productVariant

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/kafkaService"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/ShopOnGO/ShopOnGO/pkg/res"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type ProductVariantHandlerDeps struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

type ProductVariantHandler struct {
	Config *configs.Config
	Kafka  *kafkaService.KafkaService
}

func NewProductVariantHandler(router *mux.Router, deps ProductVariantHandlerDeps) {
	handler := &ProductVariantHandler{
		Config: deps.Config,
		Kafka:  deps.Kafka,
	}
	router.Handle("/product/{id}/product-variants", middleware.IsAuthed(handler.AddProductVariant(), deps.Config)).Methods("POST")
}

// AddProductVariant добавляет новый вариант продукта.
// @Summary      Добавление варианта продукта
// @Description  Добавляет новый вариант (SKU) к существующему продукту.
// @Tags         product-variant
// @Accept       json
// @Produce      json
// @Param        id    path    uint                true  "ID продукта, к которому добавляется вариант"
// @Param        body  body    addProductVariantRequest  true  "Данные нового варианта продукта (обязательно: sku, price)"
// @Success      201   {object}  map[string]interface{}  "Вариант продукта успешно создан и событие отправлено в Kafka"
// @Failure      400   {string}  string  "Неверные входные данные"
// @Failure      500   {string}  string  "Ошибка при обработке запроса"
// @Router       /product/{id}/product-variants [post]
func (h *ProductVariantHandler) AddProductVariant() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddProductVariantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		productIDStr, ok := vars["id"]
		if !ok {
			http.Error(w, "missing product ID in path", http.StatusBadRequest)
			return
		}

		productID64, err := strconv.ParseUint(productIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid product ID", http.StatusBadRequest)
			return
		}
		productID := uint(productID64)

		// Простая валидация. Можно расширить при необходимости.
		if req.SKU == "" || req.Price.LessThanOrEqual(decimal.Zero) {
			http.Error(w, "missing or invalid required fields", http.StatusBadRequest)
			return
		}
		userIDVal := r.Context().Value(middleware.ContextUserIDKey)
		var userID uint
		if id, ok := userIDVal.(uint); ok && id != 0 {
			userID = id
		}

		event := productVariantCreatedEvent{
			Action:  		"create",
			ProductID: 		productID,
			ProductVariant: req,
			UserID: 		userID,
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "failed to marshal event", http.StatusInternalServerError)
			return
		}

		if err := h.Kafka.Produce(r.Context(), []byte("product-variant-create"), eventBytes); err != nil {
			logger.Errorf("Kafka produce error: %v", err)
			http.Error(w, "failed to send Kafka message", http.StatusInternalServerError)
			return
		}

		res.Json(w, map[string]interface{}{
			"status": "product variant created and event sent",
			"data":   req,
		}, http.StatusCreated)
	}
}
