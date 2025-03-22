package di

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/brand"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
)

type IStatRepository interface {
	AddClick(linkId uint)
}

type IUserRepository interface {
	Create(user *user.User) (*user.User, error)
	FindByEmail(email string) (*user.User, error)
	Update(user *user.User) (*user.User, error)
	Delete(id uint) error
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
