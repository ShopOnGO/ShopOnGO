package di

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/product"
	"github.com/ShopOnGO/ShopOnGO/prod/internal/user"
)
import (
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
}

type IProductRepository interface {
	Create(product *product.Product) (*product.Product, error)
	GetByCategory(category *category.Category) ([]product.Product, error)
	GetByName(name string) ([]product.Product, error)
	GetFeaturedProducts(amount uint, random bool) ([]product.Product, error)
}

type ICategoryRepository interface {
	Create(product *product.Product) (*product.Product, error)
	GetFeaturedCategories(amount int) ([]category.Category, error)
	FindByName(name string) (*category.Category, error)
}

type IProductRepository interface {
	Create(product *product.Product) (*product.Product, error)
	GetByCategory(category *category.Category) ([]product.Product, error)
	GetByName(name string) ([]product.Product, error)
	GetFeaturedProducts(amount uint, random bool) ([]product.Product, error)
}

type ICategoryRepository interface {
	Create(category *category.Category) (*category.Category, error)
	GetFeaturedCategories(amount int) ([]category.Category, error)
	FindByName(name string) (*category.Category, error)
}
