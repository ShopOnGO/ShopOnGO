package cart

import "github.com/ShopOnGO/ShopOnGO/prod/pkg/db"

type CartRepository struct {
	Db *db.Db
}

func NewCartRepository(db *db.Db) *CartRepository {
	return &CartRepository{
		Db: db,
	}
}

func (repo *CartRepository) CreateCart(cart *Cart) error {
	return repo.Db.Create(cart).Error
}

func (repo *CartRepository) GetCartByID(id uint) (*Cart, error) {
	var cart Cart
	if err := repo.Db.Preload("CartItems").First(&cart, id).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (repo *CartRepository) DeleteCart(id uint) error {
	return repo.Db.Delete(&Cart{}, id).Error
}

func (repo *CartRepository) CreateCartItem(cartItem *CartItem) error {
	return repo.Db.Create(cartItem).Error
}

func (repo *CartRepository) GetCartItemsByCartID(cartID uint) ([]CartItem, error) {
	var items []CartItem
	if err := repo.Db.Where("cart_id = ?", cartID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (repo *CartRepository) DeleteCartItem(id uint) error {
	return repo.Db.Delete(&CartItem{}, id).Error
}
