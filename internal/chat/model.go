package chat

import (
	"time"
)

const (
	MsgTypeText  = "text"
	MsgTypeImage = "image"
	MsgTypeFile  = "file"
)

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FromID    uint      `gorm:"not null" json:"from_id"`
	ToID      uint      `gorm:"not null" json:"to_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      string    `gorm:"default:'text'" json:"type"`                   // "text", "image", "file"
	FileName  string    `gorm:"type:varchar(255)" json:"file_name,omitempty"` // Оригинальное имя файла (для файлов)
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
