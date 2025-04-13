package link

import (
	"github.com/ShopOnGO/ShopOnGO/pkg/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LinkRepository struct {
	Database *db.Db
}

func NewLinkRepository(database *db.Db) *LinkRepository {
	return &LinkRepository{
		Database: database,
	}
}
func (repo *LinkRepository) Create(link *Link) (*Link, error) {
	result := repo.Database.DB.Create(link)
	if result.Error != nil {
		return nil, result.Error
	}
	return link, nil
	//для создания нам не нужно указывать таблицу линк потому что мы туда передаем структуру линк,и раз он имеет горм структуру, то создается он имеено в табличке линк
	// создание, получение результата по ссылкам. это всё как обертка над обычно db только с методами
}
func (repo *LinkRepository) GetByHash(hash string) (*Link, error) {
	var link Link
	result := repo.Database.DB.First(&link, "hash = ?", hash) // SQL QUERY BY CONDS
	if result.Error != nil {
		return nil, result.Error
	}
	return &link, nil
}
func (repo *LinkRepository) Update(link *Link) (*Link, error) { // если поле в запросе не указано оно не обновляется и остается тем же
	result := repo.Database.DB.Clauses(clause.Returning{}).Updates(link)
	if result.Error != nil {
		return nil, result.Error
	}
	return link, nil
}

func (repo *LinkRepository) Delete(id uint) error {
	result := repo.Database.DB.Delete(&Link{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (repo *LinkRepository) GetById(id uint) (*Link, error) {
	var link Link                               // автоматическое lowercase и множественное число
	result := repo.Database.DB.First(&link, id) // SQL QUERY BY CONDS
	if result.Error != nil {
		return nil, result.Error
	}

	return &link, nil
}
func (repo *LinkRepository) Count() int64 {
	var count int64
	repo.Database.
		Table("links").
		Where("deleted_at is null").
		Count(&count)
	return count
}

func (repo *LinkRepository) GetAll(limit, offset int) []Link {
	var links []Link

	query := repo.Database.
		Table("links").
		Where("deleted_at is null").
		Session(&gorm.Session{})

	query.
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Scan(&links)
	return links
}
