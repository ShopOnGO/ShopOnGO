package category

import "errors"

type CategoryService struct {
	Repo *CategoryRepository
}

func NewCategoryService(repo *CategoryRepository) *CategoryService {
	return &CategoryService{Repo: repo}
}

func (s *CategoryService) CreateCategory(name, description string) (*Category, error) {
	category := &Category{
		Name:        name,
		Description: description,
	}
	return s.Repo.Create(category)
}

func (s *CategoryService) GetFeaturedCategories(amount int) ([]Category, error) {
	if amount > 20 {
		amount = 20
	}
	return s.Repo.GetFeaturedCategories(amount)
}

func (s *CategoryService) FindCategoryByName(name string) (*Category, error) {
	return s.Repo.FindByName(name)
}

func (s *CategoryService) FindCategoryByID(id uint) (*Category, error) {
	if id == 0 {
		return nil, errors.New("invalid category ID")
	}
	return s.Repo.FindCategoryByID(id)
}

func (s *CategoryService) UpdateCategory(id uint, name, description string) (*Category, error) {
	category, err := s.FindCategoryByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = description
	}

	return s.Repo.Update(category)
}

func (s *CategoryService) DeleteCategory(id uint) error {
	return s.Repo.Delete(id)
}
