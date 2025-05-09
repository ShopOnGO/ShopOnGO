package di

import (
	"time"

	"github.com/ShopOnGO/ShopOnGO/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/internal/cart"
	"github.com/ShopOnGO/ShopOnGO/internal/category"
	"github.com/ShopOnGO/ShopOnGO/internal/product"
	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/internal/user"
)

type IStatRepository interface {
	AddClick(linkId uint)
}

type IUserRepository interface {
	Create(user *user.User) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
	Update(user *user.User) (*user.User, error)
	Delete(id uint) error
	UpdateUserPassword(id uint, newPassword string) error
	GetUserRoleByEmail(email string) (string, error)
	UpdateRole(user *user.User, newRole string) error
}

type IRedisResetRepository interface { // not used on this project(redirected to notifications)
	SaveToken(email, code string, expiresAt time.Time) error
	GetToken(email string) (string, time.Time, error)
	DeleteToken(email string) error
	GetResetCodeCount(email string) (int, error)
	IncrementResetCodeCount(email string, ttl time.Duration) error
}

type IProductRepository interface {
	Create(product *product.Product) (*product.Product, error)
	GetByCategory(id uint) ([]product.Product, error)
	GetByName(name string) ([]product.Product, error)
	GetFeaturedProducts(amount uint, random bool) ([]product.Product, error)
	Update(product *product.Product) (*product.Product, error)
	Delete(id uint) error
}

type ICategoryRepository interface {
	Create(category *category.Category) (*category.Category, error)
	GetFeaturedCategories(amount int) ([]category.Category, error)
	FindByName(name string) (*category.Category, error)
	Update(category *category.Category) (*category.Category, error)
	Delete(id uint) error
}
type IBrandRepository interface {
	Create(category *brand.Brand) (*brand.Brand, error)
	GetFeaturedBrands(amount int) ([]brand.Brand, error)
	FindByName(name string) (*brand.Brand, error)
	Update(brand *brand.Brand) (*brand.Brand, error)
	Delete(id uint) error
}

type ICartRepository interface {
	GetCartByUserID(id uint) (*cart.Cart, error)
	GetCartByGuestID(guestID []byte) (*cart.Cart, error)
	GetCartItemByProductVariantID(cartID uint, productVariantID uint) (*cart.CartItem, error)
	CreateCart(cart *cart.Cart) error
	CreateCartItem(cartItem *cart.CartItem) error
	FindCartItem(cartID uint, productVariantID uint) (*cart.CartItem, error)
	UpdateCartItemQuantity(item *cart.CartItem) error
	UpdateCart(cart *cart.Cart) error
	DeleteCartItem(id uint) error
	DeleteAllCartItemsByCartID(cartID uint) error
	ClearCartItems(cartID uint) error
	DeleteCart(id uint) error
}

type IReviewRepository interface {
	CreateReview(review *review.Review) error
	GetReviewByID(id uint) (*review.Review, error)
	GetReviewsByProductVariantID(productVariantID uint) ([]review.Review, error)
	UpdateReview(review *review.Review) error
	DeleteReview(review *review.Review) error
}
