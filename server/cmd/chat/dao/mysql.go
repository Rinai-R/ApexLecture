package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/chat/model"
	"gorm.io/gorm"
)

type MysqlManagerImpl struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManagerImpl {
	db.AutoMigrate(&model.ChatMessage{})
	return &MysqlManagerImpl{db: db}
}

func (m *MysqlManagerImpl) CreateChatMessage(ctx context.Context, chatMessage *model.ChatMessage) error {
	return m.db.Create(chatMessage).Error
}
