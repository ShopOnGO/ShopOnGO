package di

import (
	"time"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/oauth2"
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
	UpdateRole(user *user.User, newRole string) (error)
}

type IURefreshTokenRepository interface {
	GetRefreshTokenData(refreshToken string) (*oauth2.RefreshTokenData, error)
	StoreRefreshToken(data *oauth2.RefreshTokenData, refreshToken string, expiresIn time.Duration) error
	DeleteRefreshToken(refreshToken string) error
	// GetUserRoleByEmail(email string) (string, error)
}

type IURedisResetRepository interface {
    SaveToken(email, code string, expiresAt time.Time) error
    GetToken(email string) (string, time.Time, error)
    DeleteToken(email string) error
}


type IProductRepository interface {
	Create(product *product.Product) (*product.Product, error)
	GetByCategory(category *category.Category) ([]product.Product, error)
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
