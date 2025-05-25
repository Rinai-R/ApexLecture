package model

import "github.com/cloudwego/eino/schema"

type Ask struct {
	History []*schema.Message `json:"history"`
	Message string            `json:"message"`
}

type AskResponse struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

// 虽然直接按照 Eino 里面规定的消息结构存储比较方便
// 但是还是自己定义一个存储历史消息的结构体便于管理
type RedisHistory struct {
	Role    string `json:"role"`
	History string `json:"history"`
}

type SummaryRequest struct {
	SummarizedText   string `json:"summarized_text"`
	UnsummarizedText string `json:"unsummarized_text"`
}

type SummaryResponse struct {
	Summary string `json:"summary"`
}

type Summary struct {
	RoomId  int64  `json:"room_id" gorm:"primary_key;type:bigint;not null"`
	Status  bool   `json:"status" gorm:"type:boolean;not null"`
	Summary string `json:"summary" gorm:"type:text;"`
}
