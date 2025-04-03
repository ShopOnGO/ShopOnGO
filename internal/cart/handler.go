package cart

import (
	"encoding/json"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/prod/configs"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/middleware"
	"github.com/gorilla/mux"
)

type CartHandlerDeps struct {
	*configs.Config
	*CartService
}
type CartHandler struct {
	*configs.Config
	*CartService
}

func NewCartHandler(router *mux.Router, deps CartHandlerDeps) {
	handler := &CartHandler{
		Config:      deps.Config,
		CartService: deps.CartService,
	}
	router.Handle("GET /cart", middleware.IsGuest(handler.GetCart(), deps.Config))
	router.Handle("POST /cart/item", middleware.IsAuthed(handler.AddCartItem(), deps.Config))
	router.Handle("PUT /cart/item", middleware.IsAuthed(handler.UpdateCartItem(), deps.Config))
	router.Handle("DELETE /cart/item", middleware.IsAuthed(handler.RemoveCartItem(), deps.Config))
	router.Handle("DELETE /cart", middleware.IsAuthed(handler.ClearCart(), deps.Config))
}

func (h *CartHandler) GetCart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cart *Cart
		var err error

		// Проверяем, есть ли userID в контексте
		if userID, ok := r.Context().Value(middleware.ContextUserIDKey).(uint); ok && userID != 0 {
			// Пользователь авторизован
			cart, err = h.CartService.GetUserCart(userID)
		} else if guestID, ok := r.Context().Value(middleware.ContextGuestIDKey).(uint); ok {
			// Пользователь — гость
			cart, err = h.CartService.GetUserCart(guestID)
		} else {
			http.Error(w, "Не удалось определить пользователя", http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, "Ошибка при получении корзины", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cart); err != nil {
			http.Error(w, "Error encoding cart response", http.StatusInternalServerError)
		}
	}
}

// Добавление товара в корзину
func (h *CartHandler) AddCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value(middleware.ContextUserIDKey).(uint)

		// Добавляем товар в корзину
		if err := h.CartService.AddItemToCart(userID, item); err != nil {
			http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// Обновление количества товара в корзине
func (h *CartHandler) UpdateCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value(middleware.ContextUserIDKey).(uint)

		// Обновляем количество товара
		if err := h.CartService.UpdateItemQuantity(userID, item); err != nil {
			http.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// Удаление товара из корзины
func (h *CartHandler) RemoveCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item CartItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value(middleware.ContextUserIDKey).(uint)

		// Удаляем товар из корзины
		if err := h.CartService.RemoveItemFromCart(userID, item); err != nil {
			http.Error(w, "Failed to remove item from cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// Очистка корзины
func (h *CartHandler) ClearCart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextUserIDKey).(uint)

		// Очищаем корзину пользователя
		if err := h.CartService.ClearUserCart(userID); err != nil {
			http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
