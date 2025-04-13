package brand

import (
	"errors"

	"github.com/ShopOnGO/ShopOnGO/prod/pkg/db"
)

type BrandRepository struct {
	Database *db.Db
}

func NewBrandRepository(database *db.Db) *BrandRepository {
	return &BrandRepository{
		Database: database,
	}
}
func (repo *BrandRepository) Create(brand *Brand) (*Brand, error) {
	result := repo.Database.DB.Create(brand)
	if result.Error != nil {
		return nil, result.Error
	}
	return brand, nil
}
func (repo *BrandRepository) GetFeaturedBrands(amount int) ([]Brand, error) {
	if amount > 20 {
		amount = 20
	}
	var brand []Brand
	query := repo.Database.DB

	if amount > 0 {
		query = query.Limit(amount)
	}

	result := query.Find(&brand)
	if result.Error != nil {
		return nil, result.Error
	}

	return brand, nil
}
func (repo *BrandRepository) FindBrandByID(id uint) (*Brand, error) {
	if id == 0 {
		return nil, errors.New("invalid brand ID")
	}
	var brand Brand
	result := repo.Database.DB.First(&brand, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &brand, nil
}
func (repo *BrandRepository) FindByName(name string) (*Brand, error) {
	var brand Brand
	result := repo.Database.DB.First(&brand, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &brand, nil
}
func (repo *BrandRepository) Update(brand *Brand) (*Brand, error) {
	if brand.ID == 0 {
		return nil, errors.New("invalid brand ID")
	}
	result := repo.Database.DB.Model(&Brand{}).Where("id = ?", brand.ID).Updates(brand)
	if result.Error != nil {
		return nil, result.Error
	}
	return brand, nil
}

func (repo *BrandRepository) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid brand ID")
	}
	result := repo.Database.DB.Delete(&Brand{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
