package category

import "gorm.io/gorm"

type Category struct {
	gorm.Model       `swaggerignore:"true"`
	Name             string     `gorm:"type:varchar(255);not null;unique" json:"name"`
	Description      string     `gorm:"type:text" json:"description"`
	ImageURL         string     `gorm:"type:varchar(255)" json:"image_url"` // Ссылка на изображение категории
	ParentCategoryID *uint      `gorm:"index"`                              // Внешний ключ может быть NULL
	ParentCategory   *Category  `gorm:"foreignKey:ParentCategoryID;constraint:OnDelete:CASCADE"`
	SubCategories    []Category `gorm:"foreignKey:ParentCategoryID"` // Связь для подкатегорий
}
