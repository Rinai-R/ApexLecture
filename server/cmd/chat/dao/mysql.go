package dao

import (
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/model"
	"gorm.io/gorm"
)

type MysqlManager struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManager {
	db.AutoMigrate(&model.ChatMessage{})
	return &MysqlManager{db: db}
}
