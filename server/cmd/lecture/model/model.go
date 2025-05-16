package model

import "time"

type Lecture struct {
	HostId      int64     `json:"host_id" gorm:"type:bigint"`
	RoomId      int64     `json:"room_id" gorm:"type:bigint;primary_key"`
	Title       string    `json:"title" gorm:"type:varchar(30);not null"`
	Description string    `json:"description" gorm:"type:varchar(100);not null"`
	Speaker     string    `json:"speaker" gorm:"type:varchar(50);not null"`
	Date        time.Time `json:"date" gorm:"type:timestamp;not null"`
}
