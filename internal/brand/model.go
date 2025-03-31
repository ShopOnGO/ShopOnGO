package brand

import "gorm.io/gorm"

type Brand struct {
	gorm.Model  `swaggerignore:"true"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	VideoURL    string `gorm:"type:varchar(255)" json:"video_url"` // Ссылка на видео в облаке
	Logo        string `gorm:"type:text" json:"logo"`              // JSON хранящий ссылку на статику(изображение)
}
