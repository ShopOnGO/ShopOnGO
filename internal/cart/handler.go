package cart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ShopOnGO/ShopOnGO/configs"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"github.com/ShopOnGO/ShopOnGO/pkg/middleware"
	"github.com/gorilla/mux"
)

type CartHandlerDeps struct {
	Config      *configs.Config
	CartService *CartService
}
type CartHandler struct {
	Config      *configs.Config
	CartService *CartService
}

func NewCartHandler(router *mux.Router, deps CartHandlerDeps) {
	handler := &CartHandler{
		Config:      deps.Config,
		CartService: deps.CartService,
	}
	router.Handle("/cart", middleware.AuthOrGuest(handler.GetCart(), deps.Config)).Methods("GET")
	router.Handle("/cart/item", middleware.AuthOrGuest(handler.AddCartItem(), deps.Config)).Methods("POST")
	router.Handle("/cart/item", middleware.AuthOrGuest(handler.UpdateCartItem(), deps.Config)).Methods("PUT")
	router.Handle("/cart/item", middleware.AuthOrGuest(handler.RemoveCartItem(), deps.Config)).Methods("DELETE")
	router.Handle("/cart", middleware.AuthOrGuest(handler.ClearCart(), deps.Config)).Methods("DELETE")
}

// GetCart returns the user's or guest's cart.
// @Summary Get Cart
// @Description Retrieves the cart for an authenticated user or guest.
// @Tags cart
// @Produce json
// @Success 200 {object} Cart "User's cart"
// @Failure 500 {string} string "Error retrieving cart"
// @Router /cart [get]
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

// AddCartItem adds a product to the user's or guest's cart.
// @Summary Add Item to Cart
// @Description Adds a product to the cart based on the request data.
// @Tags cart
// @Accept json
// @Produce json
// @Param body body AddCartItemRequest true "Product data for adding to cart"
// @Success 201 {string} string "Item successfully added to cart"
// @Failure 400 {string} string "Invalid input data"
// @Failure 500 {string} string "Error adding item to cart"
// @Router /cart/item [post]
func (h *CartHandler) AddCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Декодируем запрос в структуру AddCartItemRequest
		var req AddCartItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if req.ProductVariantID == 0 || req.Quantity <= 0 {
			http.Error(w, "Invalid ProductVariantID or Quantity", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		item := CartItem{
			ProductVariantID: req.ProductVariantID,
			Quantity:         req.Quantity,
		}

		if err := h.CartService.AddItemToCart(userID, guestID, item); err != nil {
			http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// UpdateCartItem updates the quantity of a product in the cart.
// @Summary Update Cart Item Quantity
// @Description Updates the quantity of a product in the cart for an authenticated user or guest.
// @Tags cart
// @Accept json
// @Produce json
// @Param body body UpdateCartItemRequest true "Product data for updating (must include product_variant_id and new quantity)"
// @Success 200 {string} string "Item quantity successfully updated"
// @Failure 400 {string} string "Invalid input data"
// @Failure 500 {string} string "Error updating item quantity"
// @Router /cart/item [put]
func (h *CartHandler) UpdateCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateCartItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if req.ProductVariantID == 0 || req.Quantity <= 0 {
			http.Error(w, "Invalid ProductVariantID or Quantity", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		item := CartItem{
			ProductVariantID: req.ProductVariantID,
			Quantity:         req.Quantity,
		}

		if err := h.CartService.UpdateItemQuantity(userID, guestID, item); err != nil {
			http.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// RemoveCartItem removes a product from the user's or guest's cart.
// @Summary Remove Item from Cart
// @Description Removes a product from the cart based on the data provided in the request.
// @Tags cart
// @Accept json
// @Produce json
// @Param body body RemoveCartItemRequest true "Product data for removal (must include product_variant_id)"
// @Success 200 {string} string "Item successfully removed from cart"
// @Failure 400 {string} string "Invalid input data"
// @Failure 500 {string} string "Error removing item from cart"
// @Router /cart/item [delete]
func (h *CartHandler) RemoveCartItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RemoveCartItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		if req.ProductVariantID == 0 {
			http.Error(w, "Invalid ProductVariantID", http.StatusBadRequest)
			return
		}

		userID, guestID, err := getUserOrGuestID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error(err.Error())
			return
		}

		item := CartItem{
			ProductVariantID: req.ProductVariantID,
		}

		if err := h.CartService.RemoveItemFromCart(userID, guestID, item); err != nil {
			http.Error(w, "Failed to remove item from cart", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ClearCart clears the user's or guest's cart, removing all items and the cart itself.
// @Summary Clear Cart
// @Description Clears the cart by removing all items and then the cart itself for an authenticated user or guest.
// @Tags cart
// @Produce json
// @Success 200 {string} string "Cart successfully cleared"
// @Failure 500 {string} string "Error clearing cart"
// @Router /cart [delete]
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
