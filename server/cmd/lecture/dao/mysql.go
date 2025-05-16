package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/model"
	"gorm.io/gorm"
)

type MysqlManagerImpl struct {
	DB *gorm.DB
}

func NewDM(DB *gorm.DB) *MysqlManagerImpl {
	DB.AutoMigrate(&model.Lecture{})
	return &MysqlManagerImpl{DB: DB}
}

func (m *MysqlManagerImpl) CreateLecture(ctx context.Context, lecture *model.Lecture) error {
	return m.DB.Create(lecture).Error
}
