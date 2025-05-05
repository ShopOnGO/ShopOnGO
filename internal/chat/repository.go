package chat

import (
	"github.com/ShopOnGO/ShopOnGO/pkg/db"
)

type ChatRepository struct {
	Database *db.Db
}

func NewChatRepository(database *db.Db) *ChatRepository {
	return &ChatRepository{Database: database}
}

func (r *ChatRepository) SaveMessage(message *Message) error {
	result := r.Database.DB.Create(message)
	return result.Error
}

func (r *ChatRepository) GetRecentMessages(userID uint, limit int) ([]Message, error) {
	var messages []Message
	query := r.Database.DB

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.
		Where("to_id = ? OR to_id = 0", userID).
		Order("created_at DESC").
		Find(&messages)

	return messages, result.Error
}

func (r *ChatRepository) GetLastMessages(userID uint, limit int) ([]*Message, error) {
	var messages []*Message

	err := r.Database.DB.
		Where("from_id = ? OR to_id = ?", userID, userID).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Сообщения в порядке DESC, разворачиваем в нормальный порядок (по времени)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *ChatRepository) GetMessagesBefore(userID uint, beforeMsgID uint, limit int) ([]*Message, error) {
	var messages []*Message

	err := r.Database.DB.
		Where("from_id = ? AND id < ?", userID, beforeMsgID).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Поскольку мы сортировали по DESC, надо развернуть
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
