package cart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
)

type CartHandlerDeps struct {
	*configs.Config
	*CartService
}
type CartHandler struct {
	*configs.Config
	*CartService
}

func NewCartHandler(router *http.ServeMux, deps CartHandlerDeps) {
	handler := &CartHandler{
		Config:        deps.Config,
		CartService:   deps.CartService,
	}
	router.Handle("GET /cart", middleware.AuthOrGuest(handler.GetCart(), deps.Config))
	router.Handle("POST /cart/item", middleware.AuthOrGuest(handler.AddCartItem(), deps.Config))
	router.Handle("PUT /cart/item", middleware.AuthOrGuest(handler.UpdateCartItem(), deps.Config))
	router.Handle("DELETE /cart/item", middleware.AuthOrGuest(handler.RemoveCartItem(), deps.Config))
	router.Handle("DELETE /cart", middleware.AuthOrGuest(handler.ClearCart(), deps.Config))
}

// GetCart возвращает корзину пользователя или гостя.
// @Summary      Получение корзины
// @Description  Возвращает корзину для авторизованного пользователя или гостя.
// @Tags         cart
// @Produce      json
// @Success      200  {object}  Cart   "Корзина пользователя"
// @Failure      500  {string}  string "Ошибка при получении корзины"
// @Router       /cart [get]
func (h *CartHandler) GetCart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		cart, err := h.CartService.GetCart(userID, guestID)

		if err != nil {
			http.Error(w, "Ошибка при получении корзины", http.StatusInternalServerError)
			logger.Error("Error getting cart: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cart); err != nil {
			http.Error(w, "Error encoding cart response", http.StatusInternalServerError)
			logger.Error("Error encoding cart response: ", err)
		}
	}
}

// AddCartItem добавляет товар в корзину пользователя или гостя.
// @Summary      Добавление товара в корзину
// @Description  Добавляет товар в корзину на основе данных из запроса.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        body  body  CartItem  true  "Данные товара для добавления"
// @Success      201   {string}  string  "Товар успешно добавлен"
// @Failure      400   {string}  string  "Неверные входные данные"
// @Failure      500   {string}  string  "Ошибка при добавлении товара в корзину"
// @Router       /cart/item [post]
func (h *CartHandler) AddCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		if err := h.CartService.AddItemToCart(userID, guestID, item); err != nil {
			http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// UpdateCartItem обновляет количество товара в корзине.
// @Summary      Обновление количества товара в корзине
// @Description  Обновляет количество товара в корзине для авторизованного пользователя или гостя.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        body  body  CartItem  true  "Данные товара для обновления (обязательно должен быть указан product_variant_id и новое количество)"
// @Success      200   {string}  string  "Количество товара успешно обновлено"
// @Failure      400   {string}  string  "Неверные входные данные"
// @Failure      500   {string}  string  "Ошибка при обновлении количества товара"
// @Router       /cart/item [put]
func (h *CartHandler) UpdateCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		if err := h.CartService.UpdateItemQuantity(userID, guestID, item); err != nil {
			http.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// RemoveCartItem удаляет товар из корзины пользователя или гостя.
// @Summary      Удаление товара из корзины
// @Description  Удаляет товар из корзины по данным, указанным в запросе.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        body  body  CartItem  true  "Данные товара для удаления (обязательно должен быть указан product_variant_id)"
// @Success      200   {string}  string  "Товар успешно удален из корзины"
// @Failure      400   {string}  string  "Неверные входные данные"
// @Failure      500   {string}  string  "Ошибка при удалении товара из корзины"
// @Router       /cart/item [delete]
func (h *CartHandler) RemoveCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		if err := h.CartService.RemoveItemFromCart(userID, guestID, item); err != nil {
			http.Error(w, "Failed to remove item from cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ClearCart очищает корзину пользователя или гостя, удаляя все товары и саму корзину.
// @Summary      Очистка корзины
// @Description  Очищает корзину, удаляя все товары, а затем и саму корзину для авторизованного пользователя или гостя.
// @Tags         cart
// @Produce      json
// @Success      200   {string}  string  "Корзина успешно очищена"
// @Failure      500   {string}  string  "Ошибка при очистке корзины"
// @Router       /cart [delete]
func (h *CartHandler) ClearCart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		if err := h.CartService.ClearCart(userID, guestID); err != nil {
			http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// getUserOrGuestID извлекает userID или guestID из контекста запроса.
// Если ни один из идентификаторов не найден, возвращает ошибку.
// @Summary      Извлечение идентификатора пользователя или гостя
// @Description  Извлекает из контекста запроса идентификатор авторизованного пользователя или гостя.
// @Tags         auth
// @Produce      json
// @Success      200   {object}  struct{UserID uint; GuestID []byte}  "Идентификаторы"
// @Failure      500   {string}  string  "Не удалось определить пользователя"
// @Router       /[internal] [get]
func getUserOrGuestID(r *http.Request) (uint, []byte, error) {
	var userID uint
	var guestID []byte

	if id, ok := r.Context().Value(middleware.ContextUserIDKey).(uint); ok && id != 0 {
		userID = id
	} else if id, ok := r.Context().Value(middleware.ContextGuestIDKey).([]byte); ok {
		guestID = id
	} else {
		return 0, nil, fmt.Errorf("не удалось определить пользователя: no user or guest ID in context")
	}

	return userID, guestID, nil
}
