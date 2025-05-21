package model

type ChatMessage struct {
	ID        int64  `json:"id" gorm:"primary_key;"`
	SenderID  int64  `json:"sender_id"`
	RoomID    int64  `json:"room_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
