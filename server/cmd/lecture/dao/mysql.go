package dao

import (
	"context"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/lecture/model"
	"gorm.io/gorm"
)

type MysqlManagerImpl struct {
	DB *gorm.DB
}

func NewDM(DB *gorm.DB) *MysqlManagerImpl {
	DB.AutoMigrate(&model.Lecture{})
	DB.AutoMigrate(&model.Attendance{})
	return &MysqlManagerImpl{DB: DB}
}

func (m *MysqlManagerImpl) CreateLecture(ctx context.Context, lecture *model.Lecture) error {
	return m.DB.Create(lecture).Error
}

func (m *MysqlManagerImpl) RecordJoin(ctx context.Context, attendance *model.Attendance) error {
	return m.DB.Create(attendance).Error
}

func (m *MysqlManagerImpl) RecordLeft(ctx context.Context, id int64) error {
	return m.DB.Model(&model.Attendance{}).Where("attendance_id = ?", id).Update("left_at", time.Now()).Error
}
