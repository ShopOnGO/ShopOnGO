package category

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255);not null;unique"`
	Description string `gorm:"type:text"`
	ImageURL    string `gorm:"type:varchar(255)"` // Ссылка на изображение категории
}

//TODO на выбранную базу написать формат хранения в базе и его занести в query для репозитория
