package cart

import (
	"github.com/ShopOnGO/ShopOnGO/pkg/db"
	"github.com/ShopOnGO/ShopOnGO/pkg/logger"
	"gorm.io/gorm"
)

type CartRepository struct {
	Db *db.Db
}

func NewCartRepository(db *db.Db) *CartRepository {
	return &CartRepository{
		Db: db,
	}
}

func (r *CartRepository) GetCartByUserID(userID *uint) (*Cart, error) {
	var cart Cart
	if err := r.Db.
    Preload("CartItems", func(db *gorm.DB) *gorm.DB {
        return db.Preload("ProductVariant")
    }).
    Where("user_id = ?", userID).
    First(&cart).Error; err != nil {
    	return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) GetCartByGuestID(guestID []byte) (*Cart, error) {
	var cart Cart

	if err := r.Db.
    Preload("CartItems", func(db *gorm.DB) *gorm.DB {
        return db.Preload("ProductVariant")
		//.Preload("ProductVariant.Images") // можно еще например
    }).
    Where("guest_id = ?", []byte(guestID)).
    First(&cart).Error; err != nil {
    	return nil, err
	}

	logger.Infof("Cart found: %+v", cart)
	return &cart, nil
}

func (r *CartRepository) GetCartItemByProductVariantID(cartID uint, productVariantID uint) (*CartItem, error) {
	var item CartItem
	err := r.Db.Where("cart_id = ? AND product_variant_id = ?", cartID, productVariantID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) CreateCart(cart *Cart) error {
	return r.Db.Create(cart).Error
}

func (r *CartRepository) CreateCartItem(item *CartItem) error {
	return r.Db.Create(item).Error
}

func (r *CartRepository) FindCartItem(cartID uint, productVariantID uint) (*CartItem, error) {
	var item CartItem
	err := r.Db.Where("cart_id = ? AND product_variant_id = ?", cartID, productVariantID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) UpdateCartItemQuantity(item *CartItem) error {
    return r.Db.Model(item).Update("quantity", item.Quantity).Error
}

func (r *CartRepository) UpdateCart(cart *Cart) error {
    return r.Db.Save(cart).Error
}

func (r *CartRepository) DeleteCartItem(itemID uint, cartID uint) error {
	return r.Db.Where("id = ? AND cart_id = ?", itemID, cartID).Delete(&CartItem{}).Error
}

func (r *CartRepository) DeleteAllCartItemsByCartID(cartID uint) error {
    return r.Db.Where("cart_id = ?", cartID).Delete(&CartItem{}).Error
}

func (r *CartRepository) ClearCartItems(cartID uint) error {
	return r.Db.Where("cart_id = ?", cartID).Delete(&CartItem{}).Error
}

func (r *CartRepository) DeleteCart(cartID uint) error {
	return r.Db.Delete(&Cart{}, cartID).Error
}