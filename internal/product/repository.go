package product

import (
	"errors"

	"github.com/ShopOnGO/ShopOnGO/internal/review"
	"github.com/ShopOnGO/ShopOnGO/pkg/db"
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
	if product.Name == "" || product.CategoryID == 0 {
		return nil, errors.New("product name and category ID are required")
	}
	result := repo.Database.DB.Create(product)
	if result.Error != nil {
		return nil, result.Error
	}
	return product, nil
}

func (repo *ProductRepository) GetByCategory(CategoryID uint) ([]Product, error) { //limit 20
	if CategoryID == 0 {
		return nil, errors.New("category cannot be nil")
	}
	var products []Product
	result := repo.Database.DB.
		Where("category_id = ?", CategoryID).
		Limit(20).
		Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// func (repo *ProductRepository) GetByVendorCode(code *uuid)
func (repo *ProductRepository) GetByName(name string) ([]Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}
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
	if amount == 0 {
		return nil, errors.New("amount must be greater than zero")
	}
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

func (repo *ProductRepository) GetProductWithReviewsByID(productID uint) (*Product, error) {
	if productID == 0 {
		return nil, errors.New("product ID is required")
	}

	var product Product

	// Шаг 1: Загружаем продукт с вариантами
	err := repo.Database.DB.
		Preload("Variants").
		First(&product, productID).
		Error
	if err != nil {
		return nil, err
	}

	// Шаг 2: Собираем все Variant.ID
	variantIDs := make([]uint, 0, len(product.Variants))
	for _, v := range product.Variants {
		variantIDs = append(variantIDs, v.ID)
	}

	// Шаг 3: Загружаем отзывы по этим variant_id
	var reviews []review.Review
	err = repo.Database.DB.
		Where("product_variant_id IN ?", variantIDs).
		Preload("User"). // если хочешь подгружать пользователей
		Find(&reviews).Error
	if err != nil {
		return nil, err
	}

	product.Reviews = reviews
	return &product, nil
}


func (repo *ProductRepository) Update(product *Product) (*Product, error) {
	if product.ID == 0 {
		return nil, errors.New("product ID is required for update")
	}
	result := repo.Database.DB.Model(&Product{}).Where("id = ?", product.ID).Updates(product)
	if result.Error != nil {
		return nil, result.Error
	}
	return product, nil
}

func (repo *ProductRepository) Delete(id uint) error {
	if id == 0 {
		return errors.New("product ID is required for deletion")
	}
	result := repo.Database.DB.Delete(&Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
