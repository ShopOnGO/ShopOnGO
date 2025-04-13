package cart

import (
	"fmt"

	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"gorm.io/gorm"
)

type CartService struct {
	Repo *CartRepository
}

func NewCartService(repo *CartRepository) *CartService {
	return &CartService{
		Repo: repo,
	}
}

func (s *CartService) GetCart(userID *uint, guestID []byte) (*Cart, error) {
	if userID != nil {
		if len(guestID) > 0 {
            if err := s.MergeCarts(userID, guestID); err != nil {
                logger.Error("Ошибка при объединении корзин: ", err)
            }
        }
		cart, err := s.Repo.GetCartByUserID(userID)
		if err == nil {
			return cart, nil
		}
		newCart := &Cart{UserID: userID}
		if err = s.Repo.CreateCart(newCart); err != nil {
			return nil, fmt.Errorf("failed to create user cart: %w", err)
		}
		return newCart, nil
	}

	if len(guestID) > 0 {
		cart, err := s.Repo.GetCartByGuestID(guestID)
		if err == nil {
			return cart, nil
		}
		newCart := &Cart{GuestID: guestID}
		if err = s.Repo.CreateCart(newCart); err != nil {
			return nil, fmt.Errorf("failed to create guest cart: %w", err)
		}
		return newCart, nil
	}

	return nil, fmt.Errorf("no valid userID or guestID provided")
}

func (s *CartService) AddItemToCart(userID *uint, guestID []byte, item CartItem) error {
	cart, err := s.GetCart(userID, guestID)
	if err != nil {
		return err
	}
	existingItem, err := s.Repo.GetCartItemByProductVariantID(cart.ID, item.ProductVariantID)
	if err == nil {
		existingItem.Quantity += item.Quantity
		if err := s.Repo.UpdateCartItemQuantity(existingItem); err != nil {
			return fmt.Errorf("failed to update item quantity: %w", err)
		}
		return nil
	}

	item.CartID = cart.ID
	if err := s.Repo.CreateCartItem(&item); err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	return nil
}

func (s *CartService) UpdateItemQuantity(userID *uint, guestID []byte, item CartItem) error {
	cart, err := s.GetCart(userID, guestID)
	if err != nil {
		logger.Error("failed to get cart: ", err)
		return fmt.Errorf("failed to get cart: %w", err)
	}

	existingItem, err := s.Repo.FindCartItem(cart.ID, item.ProductVariantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("item not found in cart")
			return fmt.Errorf("item not found in cart")
		}
		logger.Error("failed to find item in cart: ", err)
		return fmt.Errorf("failed to find item in cart: %w", err)
	}

	existingItem.Quantity = item.Quantity
	if err := s.Repo.UpdateCartItemQuantity(existingItem); err != nil {
		logger.Error("failed to update item quantity")
		return fmt.Errorf("failed to update item quantity: %w", err)
	}

	return nil
}

func (s *CartService) RemoveItemFromCart(userID *uint, guestID []byte, item CartItem) error {
	cart, err := s.GetCart(userID, guestID)
	if err != nil {
		return err
	}

	existingItem, err := s.Repo.FindCartItem(cart.ID, item.ProductVariantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Error("item not found in cart")
			return fmt.Errorf("item not found in cart")
		}
		logger.Error("failed to find item in cart: ", err)
		return fmt.Errorf("failed to find item in cart: %w", err)
	}

	if err := s.Repo.DeleteCartItem(existingItem.ID, cart.ID); err != nil {
		return fmt.Errorf("failed to remove item from cart: %w", err)
	}

	return nil
}

func (s *CartService) ClearCart(userID *uint, guestID []byte) error {
	cart, err := s.GetCart(userID, guestID)
	if err != nil {
		return err
	}

	if err := s.Repo.ClearCartItems(cart.ID); err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	if err := s.Repo.DeleteCart(cart.ID); err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	return nil
}


func (s *CartService) MergeCarts(userID *uint, guestID []byte) error {
    if userID == nil || len(guestID) == 0 {
        return nil
    }
    
    guestCart, err := s.Repo.GetCartByGuestID(guestID)
    if err != nil {
        return nil
    }
    
    userCart, err := s.Repo.GetCartByUserID(userID)
    if err != nil {
        guestCart.UserID = userID
        guestCart.GuestID = nil
        return s.Repo.UpdateCart(guestCart)
    }
    
    for _, guestItem := range guestCart.CartItems {
        existingItem, err := s.Repo.GetCartItemByProductVariantID(userCart.ID, guestItem.ProductVariantID)
        if err == nil {
            existingItem.Quantity += guestItem.Quantity
            if err := s.Repo.UpdateCartItemQuantity(existingItem); err != nil {
                return fmt.Errorf("failed to update item quantity: %w", err)
            }
        } else {
            guestItem.CartID = userCart.ID
            if err := s.Repo.CreateCartItem(&guestItem); err != nil {
                return fmt.Errorf("failed to create cart item: %w", err)
            }
        }
    }

	if err := s.Repo.DeleteAllCartItemsByCartID(guestCart.ID); err != nil {
        return fmt.Errorf("failed to delete guest cart items: %w", err)
    }
    
    return s.Repo.DeleteCart(guestCart.ID)
}