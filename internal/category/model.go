package category

import "gorm.io/gorm"

type Category struct {
	gorm.Model  `swaggerignore:"true"`
	Name        string `gorm:"type:varchar(255);not null;unique" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ImageURL    string `gorm:"type:varchar(255)" json:"image_url"` // Ссылка на изображение категории
}
