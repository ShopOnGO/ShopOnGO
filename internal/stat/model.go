package stat

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Stat struct {
	gorm.Model `swaggerignore:"true"`
	LinkId     uint           `json:"link_id"`
	Clicks     int            `json:"clicks"`
	Date       datatypes.Date `json:"date" swaggertype:"string" format:"date"` // поддерживается в postgres
}
