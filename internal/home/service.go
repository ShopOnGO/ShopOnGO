package home

import (
	"github.com/ShopOnGO/ShopOnGO/pkg/di"
)

type HomeService struct {
	CategoryRepository di.ICategoryRepository
	BrandRepository    di.IBrandRepository
	//promoRepo PromotionRepository
}

func NewHomeService(categoryRepository di.ICategoryRepository, brandRepository di.IBrandRepository) *HomeService {
	return &HomeService{
		CategoryRepository: categoryRepository,
		BrandRepository:    brandRepository}
}
func (s *HomeService) GetHomeData() (*HomeData, error) {

	categories, err := s.CategoryRepository.GetFeaturedCategories(5)
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
		Brands:           brands,
		//Promotions: promotions,
	}, nil
}
