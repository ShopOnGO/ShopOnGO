package user

import "github.com/ShopOnGO/ShopOnGO/prod/pkg/db"

type UserRepository struct {
	Database *db.Db
}

func NewUserRepository(database *db.Db) *UserRepository {
	return &UserRepository{
		Database: database,
	}
}

func (repo *UserRepository) Create(user *User) (*User, error) {
	result := repo.Database.DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
	//для создания нам не нужно указывать таблицу линк потому что мы туда передаем структуру линк,и раз он имеет горм структуру, то создается он имеено в табличке линк
	// создание, получение результата по ссылкам. это всё как обертка над обычно db только с методами
}

func (repo *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	result := repo.Database.DB.First(&user, "email = ?", email) // SQL QUERY BY CONDS
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (repo *UserRepository) Update(user *User) (*User, error) {
	result := repo.Database.DB.Model(&User{}).Where("id = ?", user.ID).Updates(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (repo *UserRepository) Delete(id uint) error {
	result := repo.Database.DB.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
