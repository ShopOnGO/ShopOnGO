package brand

import "errors"

type BrandService struct {
	Repo *BrandRepository
}

func NewBrandService(repo *BrandRepository) *BrandService {
	return &BrandService{Repo: repo}
}

func (s *BrandService) CreateBrand(name, description, videoURL, logo string) (*Brand, error) {
	brand := &Brand{
		Name:        name,
		Description: description,
		VideoURL:    videoURL,
		Logo:        logo,
	}
	return s.Repo.Create(brand)
}

func (s *BrandService) GetFeaturedBrands(amount int) ([]Brand, error) {
	if amount > 20 {
		amount = 20
	}
	return s.Repo.GetFeaturedBrands(amount)
}

func (s *BrandService) FindBrandByName(name string) (*Brand, error) {
	return s.Repo.FindByName(name)
}

func (s *BrandService) FindBrandByID(id uint) (*Brand, error) {
	if id == 0 {
		return nil, errors.New("invalid brand ID")
	}
	return s.Repo.FindBrandByID(id)
}

func (s *BrandService) UpdateBrand(id uint, name, description, videoURL, logo string) (*Brand, error) {

	brand, err := s.FindBrandByID(id)
	if err != nil {
		return nil, err
	}

	// searching active fields
	if name != "" {
		brand.Name = name
	}
	if description != "" {
		brand.Description = description
	}
	if videoURL != "" {
		brand.VideoURL = videoURL
	}
	if logo != "" {
		brand.Logo = logo
	}

	// 3. Сохранить изменения
	return s.Repo.Update(brand)
}

func (s *BrandService) DeleteBrand(id uint) error {
	return s.Repo.Delete(id)
}
