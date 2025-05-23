package dao

import "gorm.io/gorm"

type MysqlManagerImpl struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManagerImpl {
	return &MysqlManagerImpl{db: db}
}
