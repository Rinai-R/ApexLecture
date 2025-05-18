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

type Attendance struct {
	AttendanceId int64     `json:"attendance_id" gorm:"type:bigint;primary_key"`
	RoomId       int64     `json:"room_id" gorm:"type:bigint;not null"`
	UserId       int64     `json:"user_id" gorm:"type:bigint;not null"`
	JoinAt       time.Time `json:"join_at" gorm:"type:timestamp;not null;default:current_timestamp"`
	LeftAt       time.Time `json:"left_at" gorm:"type:timestamp;not null;default:current_timestamp"`
}
