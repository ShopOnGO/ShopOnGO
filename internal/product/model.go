package product

import "gorm.io/gorm"

type Product struct {
	gorm.Model  `swaggerignore:"true"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	//CategoryID  uint   `gorm:"not null" json:"category_id"`//foreign key
	//BrandID      uint    `gorm:"not null"  json:"brand_id"`
	//Price        float64 `gorm:"not null"  json:"price"`
	Discount    float64 `gorm:"default:0" json:"discount"`
	Stock       int     `gorm:"not null;default:0" json:"stock"`  // количество в наличии
	IsAvailable bool    `gorm:"default:true" json:"is_available"` // доступен
	Size        string  `gorm:"type:varchar(50)" json:"size"`
	Color       string  `gorm:"type:varchar(50)" json:"color"`
	Material    string  `gorm:"type:varchar(100)" json:"material"`
	Gender      string  `gorm:"type:varchar(20)" json:"gender"`
	Season      string  `gorm:"type:varchar(20)" json:"season"`
	//ImageURL    string  `gorm:"type:varchar(255)" json:"image_url"`
	VideoURL string `gorm:"type:varchar(255)" json:"video_url"` // Ссылка на видео в облаке
	Gallery  string `gorm:"type:text" json:"gallery"`           // JSON хранящий ссылки на изображения
	//VendorCode   string  `gorm:"type:varchar(100);unique;not null"json:"vendor_code"`//артикул
	Rating       float64 `gorm:"default:0" json:"rating"`
	ReviewsCount int     `gorm:"default:0" json:"reviews_count"` // количество отзывов
	//popularity
}

//todo category_id (3)
