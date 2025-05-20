package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/model"
	"gorm.io/gorm"
)

type MysqlManager struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManager {
	db.AutoMigrate(&model.Message{})
	db.AutoMigrate(&model.UserAnswer{})
	return &MysqlManager{db: db}
}

func (m *MysqlManager) CreateRoom(ctx context.Context, room *model.Room) error {
	return m.db.WithContext(ctx).Create(room).Error
}
