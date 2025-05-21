package model

import "time"

type ChatMessage struct {
	ID        int64     `json:"id" gorm:"primary_key;type:bigint;"`
	SenderID  int64     `json:"sender_id" gorm:"not null;type:bigint;"`
	RoomID    int64     `json:"room_id" gorm:"not null;type:bigint;"`
	Content   string    `json:"content" gorm:"not null;type:text;"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;type:timestamp;default:CURRENT_TIMESTAMP;"`
}
