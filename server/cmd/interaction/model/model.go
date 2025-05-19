package model

import "time"

type Message struct {
	ID          int64         `json:"id" gorm:"column:id;primary_key;not null"`
	RoomID      int64         `json:"room_id" gorm:"column:room_id;not null"`
	UserID      int64         `json:"user_id" gorm:"column:user_id;not null"`
	Content     string        `json:"content" gorm:"column:content;not null"`
	CreatedAt   time.Time     `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	VideoOffset time.Duration `json:"video_offset" gorm:"column:video_offset;not null;type:bigint"`
}

type TrueFalse struct {
	ID          int64         `json:"id" gorm:"column:id;primary_key;not null"`
	RoomID      int64         `json:"room_id" gorm:"column:room_id;not null"`
	Title       string        `json:"title" gorm:"column:title;not null"`
	Answer      bool          `json:"answer" gorm:"column:answer;not null"`
	VideoOffset time.Duration `json:"video_offset" gorm:"column:video_offset;not null;type:bigint"`
	CreatedAt   time.Time     `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

type Choice struct {
	ID          int64         `json:"id" gorm:"column:id;primary_key;not null"`
	RoomID      int64         `json:"room_id" gorm:"column:room_id;not null"`
	Title       string        `json:"title" gorm:"column:title;not null"`
	Answer      string        `json:"answer" gorm:"column:answer;not null"`
	VideoOffset time.Duration `json:"video_offset" gorm:"column:video_offset;not null;type:bigint"`
}

type Text struct {
	ID          int64         `json:"id" gorm:"column:id;primary_key;not null"`
	RoomID      int64         `json:"room_id" gorm:"column:room_id;not null"`
	Content     string        `json:"content" gorm:"column:content;not null"`
	VideoOffset time.Duration `json:"video_offset" gorm:"column:video_offset;not null;type:bigint"`
}
