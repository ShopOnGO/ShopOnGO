package link

import (
	"github.com/ShopOnGO/ShopOnGO/internal/stat"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model `swaggerignore:"true"`
	Url        string      `json:"url"`
	Hash       string      `json:"hash" gorm:"uniqueIndex"`
	Stats      []stat.Stat `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	//поставили каскадную связь между таблицами что не позволит просто так удалить ссылку, так как она может относиться ко множеству статистик
	//ограничения некритичны
}
