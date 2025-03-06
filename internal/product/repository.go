package product

import (
	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
)

type ProductRepository struct {
	Database *db.Db
}

func NewProductRepository(database *db.Db) *ProductRepository {
	return &ProductRepository{
		Database: database,
	}
}

func (repo *ProductRepository) Create(product *Product) (*Product, error) {
	result := repo.Database.DB.Create(product)
	if result.Error != nil {
		return nil, result.Error
	}
	return product, nil
}

func (repo *ProductRepository) GetByCategory(category *category.Category) ([]Product, error) { //limit 20
	var products []Product
	result := repo.Database.DB.
		Where("category_id = ?", category.ID).
		Limit(20).
		Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// func (repo *ProductRepository) GetByVendorCode(code *uuid)
func (repo *ProductRepository) GetByName(name string) ([]Product, error) {
	var products []Product
	result := repo.Database.DB.
		Where("name = ?", name).
		Limit(20).
		Find(&products) // table should be named "products"
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

func (repo *ProductRepository) GetFeaturedProducts(amount uint, random bool) ([]Product, error) {
	var products []Product
	query := repo.Database.DB

	if random {
		query = query.Order("RAND()")
	} else {
		// Получение товаров по популярности НЕ РАБОТАЕТ ЕЩЕ!!
		query = query.Order("popularity DESC")
	}

	result := query.Limit(int(amount)).Find(&products).Where("deleted_at is null")

	return products, result.Error
}
