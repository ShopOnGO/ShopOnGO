package productVariant

import (
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)
		
type ProductVariant struct {
	gorm.Model
	ProductID		uint      			`gorm:"index;not null"`                // на всякий
	SKU           	string    			`gorm:"type:varchar(100);uniqueIndex"` // Уникальный артикул
	Price    	  	decimal.Decimal 	`gorm:"type:decimal(8,2);not null"`
	Discount 	  	decimal.Decimal 	`gorm:"type:decimal(8,2);not null;default:0"`
	ReservedStock 	uint32    			`gorm:"not null"` // бронь (пока оплатишь типа)
	Rating        	decimal.Decimal 	`gorm:"type:decimal(8,1);not null;default:0"`
	ReviewCount   	uint      			`gorm:"not null;default:0"`
	RatingSum     	uint	  			`gorm:"not null;default:0"`
	Sizes  			string 				`gorm:"type:varchar(255)" json:"sizes"`
	Colors 			string 				`gorm:"type:varchar(255)" json:"colors"`
	Stock         	uint32    			`gorm:"default:0"`         // Общий остаток на складе
	//Weight          uint      		`gorm:"default:0"`         // Вес в граммах
	Barcode    	  	string   			`gorm:"type:varchar(50)"`  // Штрих-код
	IsActive   		bool     			`gorm:"default:true"`      // Активен ли вариант
	ImageURLs 		pq.StringArray 		`gorm:"type:text[]"`
	MinOrder   		uint     			`gorm:"default:1"`         // Минимальный заказ
	Dimensions 		string   			`gorm:"type:varchar(50)"`  // Габариты (например "20x30x5 см")
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
