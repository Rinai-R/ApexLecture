package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/model"
	"gorm.io/gorm"
)

type DM struct {
	db *gorm.DB
}

func NewDM(db *gorm.DB) *DM {
	err := db.AutoMigrate(&model.User{})
	if err != nil {
		panic("failed to migrate user table " + err.Error())
	}
	return &DM{db: db}
}

func (d *DM) CreateUser(ctx context.Context, user *model.User) error {
	res := d.db.Create(user)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *DM) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := d.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
