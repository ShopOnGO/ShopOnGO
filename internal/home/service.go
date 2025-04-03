package home

import (
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/di"
)

type HomeService struct {
	CategoryRepository di.ICategoryRepository
	ProductsRepository di.IProductRepository
	BrandRepository    di.IBrandRepository
	//promoRepo PromotionRepository
}

func NewHomeService(categoryRepository di.ICategoryRepository, productsRepository di.IProductRepository, brandRepository di.IBrandRepository) *HomeService {
	return &HomeService{
		CategoryRepository: categoryRepository,
		ProductsRepository: productsRepository,
		BrandRepository:    brandRepository}
}
func (s *HomeService) GetHomeData() (*HomeData, error) {

	categories, err := s.CategoryRepository.GetFeaturedCategories(5)
	if err != nil {
		return nil, err
	}

	featuredProducts, err := s.ProductsRepository.GetFeaturedProducts(10, true) // ONLY TRUE WHILE POPULARITY IS UNDEF
	if err != nil {
		return nil, err
	}

	// promotions, err := s.promoRepo.GetActivePromotions()
	// if err != nil {
	// 	return nil, err
	// }
	brands, err := s.BrandRepository.GetFeaturedBrands(5)
	if err != nil {
		return nil, err
	}

	return &HomeData{
		Categories:       categories,
		FeaturedProducts: featuredProducts,
		Brands:           brands,
		//Promotions: promotions,
	}, nil
}
