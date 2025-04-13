package productVariant

import (
	"gorm.io/gorm"
)

type ProductVariant struct {
	gorm.Model
	ProductID     uint     `gorm:"index;not null"`                // на всякий
	SKU           string   `gorm:"type:varchar(100);uniqueIndex"` // Уникальный артикул
	Price         uint32   `gorm:"type:decimal(8,2);not null"`
	Discount      uint32   `gorm:"type:decimal(8,2);not null;default:0"`
	ReservedStock uint32   `gorm:"not null"` // бронь (пока оплатишь типа)
	Rating        uint     `gorm:"not null;default:0"`
	Sizes         []uint32 `gorm:"type:json"`         // Храним размеры как JSON-массив
	Colors        []string `gorm:"type:json"`         // Храним цвета как JSON-массив
	Stock         uint32   `gorm:"default:0"`         // Общий остаток на складе
	Material      string   `gorm:"type:varchar(200)"` // Материал изготовления
	//Weight          uint      `gorm:"default:0"`                           // Вес в граммах
	Barcode    string   `gorm:"type:varchar(50)"` // Штрих-код
	IsActive   bool     `gorm:"default:true"`     // Активен ли вариант
	Images     []string `gorm:"type:json"`        // Массив URL изображений
	MinOrder   uint     `gorm:"default:1"`        // Минимальный заказ
	Dimensions string   `gorm:"type:varchar(50)"` // Габариты (например "20x30x5 см")
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
