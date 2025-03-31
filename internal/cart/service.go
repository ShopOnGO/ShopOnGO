package cart

import (
	"fmt"
)

// CartService описывает бизнес-логику для работы с корзиной.
type CartService struct {
	Repo *CartRepository
}

// NewCartService создаёт новый экземпляр CartService.
func NewCartService(repo *CartRepository) *CartService {
	return &CartService{
		Repo: repo,
	}
}

// GetUserCart получает корзину для пользователя по его идентификатору.
// Если корзина не существует, создаётся новая.
func (s *CartService) GetUserCart(userID uint) (*Cart, error) {
	// Преобразуем userID в тип uint, если в БД идентификатор хранится как uint.
	cart, err := s.Repo.GetCartByID(userID)
	if err != nil {
		// Если корзина не найдена, создаём новую
		newCart := &Cart{
			UserID: uint(userID),
		}
		if err = s.Repo.CreateCart(newCart); err != nil {
			return nil, fmt.Errorf("failed to create new cart: %w", err)
		}
		return newCart, nil
	}
	return cart, nil
}

// AddItemToCart добавляет элемент в корзину пользователя.
func (s *CartService) AddItemToCart(userID uint, item CartItem) error {
	// Получаем корзину пользователя. Если её нет – создаётся новая.
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return err
	}
	// Привязываем элемент к корзине.
	item.CartID = cart.ID
	if err := s.Repo.CreateCartItem(&item); err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}
	return nil
}

// UpdateItemQuantity обновляет количество товара в корзине пользователя.
// Предполагается, что CartItem содержит поля ID, CartID и Quantity.
func (s *CartService) UpdateItemQuantity(userID uint, item CartItem) error {
	// Получаем корзину пользователя.
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return err
	}
	// Проверяем, что элемент принадлежит корзине пользователя.
	if item.CartID != cart.ID {
		return fmt.Errorf("item does not belong to user's cart")
	}
	// Обновляем количество товара. Здесь мы напрямую используем доступ к базе через репозиторий.
	// Предполагается, что s.Repo.Db соответствует *db.Db, обёртке над GORM.
	if err := s.Repo.Db.Model(&CartItem{}).
		Where("id = ? AND cart_id = ?", item.ID, cart.ID).
		Update("quantity", item.Quantity).Error; err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}
	return nil
}

// RemoveItemFromCart удаляет элемент из корзины пользователя.
func (s *CartService) RemoveItemFromCart(userID uint, item CartItem) error {
	// Получаем корзину пользователя.
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return err
	}
	// Проверяем, что элемент действительно принадлежит корзине.
	if item.CartID != cart.ID {
		return fmt.Errorf("item does not belong to user's cart")
	}
	// Удаляем элемент, используя репозиторий.
	if err := s.Repo.DeleteCartItem(item.ID); err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}
	return nil
}

// ClearUserCart очищает корзину пользователя, удаляя все товары, привязанные к ней.
func (s *CartService) ClearUserCart(userID uint) error {
	cart, err := s.GetUserCart(userID)
	if err != nil {
		return err
	}
	// Удаляем все товары, связанные с корзиной.
	if err := s.Repo.Db.Where("cart_id = ?", cart.ID).Delete(&CartItem{}).Error; err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}