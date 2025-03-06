package category

import (
	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
)

type CategoryRepository struct {
	Database *db.Db
}

func NewUserRepository(database *db.Db) *CategoryRepository {
	return &CategoryRepository{
		Database: database,
	}
}
func (repo *CategoryRepository) Create(category *Category) (*Category, error) {
	result := repo.Database.DB.Create(category)
	if result.Error != nil {
		return nil, result.Error
	}
	return category, nil
}
func (repo *CategoryRepository) GetCategories() ([]Category, error) {
	var categories []Category

	result := repo.Database.DB.Find(&categories) // table should be named categories
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}
func (repo *CategoryRepository) FindByName(name string) (*Category, error) {
	var category Category
	result := repo.Database.DB.First(&category, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &category, nil
}
