package product

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	//CategoryID   uint    `gorm:"not null"`
	//BrandID      uint    `gorm:"not null"`
	//Price        float64 `gorm:"not null"`
	Discount    float64 `gorm:"default:0"`
	Stock       int     `gorm:"not null;default:0"` // количество в наличии
	IsAvailable bool    `gorm:"default:true"`       // доступен
	Size        string  `gorm:"type:varchar(50)"`
	Color       string  `gorm:"type:varchar(50)"`
	Material    string  `gorm:"type:varchar(100)"`
	Gender      string  `gorm:"type:varchar(20)"`
	Season      string  `gorm:"type:varchar(20)"`
	ImageURL    string  `gorm:"type:varchar(255)"`
	VideoURL    string  `gorm:"type:varchar(255)"` // Ссылка на видео в облаке
	Gallery     string  `gorm:"type:text"`         // JSON хранящий ссылки на изображения
	//VendorCode   string  `gorm:"type:varchar(100);unique;not null"`//артикул
	Rating       float64 `gorm:"default:0"`
	ReviewsCount int     `gorm:"default:0"` // количество отзывов
}
