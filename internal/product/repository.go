package product

import (
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

func (repo *ProductRepository) GetByCategory(id uint) ([]Product, error) { //limit 20
	var products []Product
	result := repo.Database.DB.
		Where("category_id = ?", id).
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
		query = query.Order("RANDOM()")
	} else {
		// Получение товаров по популярности НЕ РАБОТАЕТ ЕЩЕ!!
		query = query.Order("popularity DESC")
	}

	result := query.Limit(int(amount)).Find(&products).Where("deleted_at is null")

	return products, result.Error
}

func (repo *ProductRepository) Update(product *Product) (*Product, error) {
	result := repo.Database.DB.Model(&Product{}).Where("id = ?", product.ID).Updates(product)
	if result.Error != nil {
		return nil, result.Error
	}
	return product, nil
}

func (repo *ProductRepository) Delete(id uint) error {
	result := repo.Database.DB.Delete(&Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
