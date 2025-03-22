package product

import (
	"errors"

	"github.com/ShopOnGO/ShopOnGO/prod/internal/category"
)

type ProductService struct {
	ProductRepository *ProductRepository
}

func NewProductService(productRepository *ProductRepository) *ProductService {
	return &ProductService{ProductRepository: productRepository}
}

func (s *ProductService) CreateProduct(product *Product) (*Product, error) {
	if product.Name == "" || product.CategoryID == 0 {
		return nil, errors.New("product name and category ID are required")
	}
	return s.ProductRepository.Create(product)
}

func (s *ProductService) GetProductsByCategory(category *category.Category) ([]Product, error) {
	if category == nil {
		return nil, errors.New("category cannot be nil")
	}
	return s.ProductRepository.GetByCategory(category.ID)
}

func (s *ProductService) GetProductsByName(name string) ([]Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	return s.ProductRepository.GetByName(name)
}

func (s *ProductService) GetFeaturedProducts(amount uint, random bool) ([]Product, error) {
	if amount == 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	return s.ProductRepository.GetFeaturedProducts(amount, random)
}

func (s *ProductService) UpdateProduct(product *Product) (*Product, error) {
	if product.ID == 0 {
		return nil, errors.New("product ID is required for update")
	}
	return s.ProductRepository.Update(product)
}

func (s *ProductService) DeleteProduct(id uint) error {
	if id == 0 {
		return errors.New("product ID is required for deletion")
	}
	return s.ProductRepository.Delete(id)
}
