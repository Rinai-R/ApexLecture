package dao

import (
	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/model"
	"gorm.io/gorm"
)

type MysqlManager struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManager {
	db.AutoMigrate(&model.Choice{})
	db.AutoMigrate(&model.Text{})
	db.AutoMigrate(&model.Message{})
	db.AutoMigrate(&model.TrueFalse{})
	return &MysqlManager{db: db}
}
