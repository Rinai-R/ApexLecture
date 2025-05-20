package model

import (
	"time"
)

type Room struct {
	RoomID    int64     `json:"room_id" gorm:"column:room_id;primary_key;unique;not null"`
	HostID    int64     `json:"host_id" gorm:"column:host_id;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
}

const (
	MessageTypeChat      int8 = iota + 1 // 普通聊天消息
	MessageTypeTrueFalse                 // 判断题
	MessageTypeChoice                    // 选择题
	MessageTypeText                      // 文本题
)

// Message 统一的消息结构
// 包括消息 ID，房间 ID，发送者 ID，消息类型，消息内容/题目，题目附带的信息，视频偏移量，创建时间
type Message struct {
	ID          int64         `json:"id" gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;not null"`
	RoomID      int64         `json:"room_id" gorm:"column:room_id;type:bigint unsigned;not null;index"`
	SenderID    int64         `json:"sender_id" gorm:"column:sender_id;type:bigint unsigned;not null;index"`
	Type        int8          `json:"type" gorm:"column:type;type:tinyint unsigned;not null"`
	Content     string        `json:"content" gorm:"column:content;type:text;not null"` // 聊天内容或题目标题
	Extra       MessageExtra  `json:"extra" gorm:"column:extra;type:json"`              // 题目的额外信息（选项、答案等）
	VideoOffset time.Duration `json:"video_offset" gorm:"column:video_offset;type:bigint;not null"`
	CreatedAt   time.Time     `json:"created_at" gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// MessageExtra 额外字段
// 这里是因为不同的题目有着不同的字段
type MessageExtra struct {
	Options         []string `json:"options,omitempty"`
	Score           float64  `json:"score,omitempty"`
	AnswerText      string   `json:"answer_text,omitempty"`
	AnswerChoice    []int8   `json:"answer_choice,omitempty"`
	AnswerTrueFalse bool     `json:"answer_true_false,omitempty"`
}

// 体验一下 WA 的感受😎
const (
	Judging     int8 = iota // 待评分
	Accept                  // 正确
	WrongAnswer             // 错误
	Judged                  // 已评分
)

// UserAnswer 用户答题记录
// 包括答案 ID，用户 ID，消息 ID，答案，批改状态，得分，评语，创建时间，更新时间
type UserAnswer struct {
	ID        int64     `json:"id" gorm:"column:id;type:bigint unsigned;primary_key;auto_increment;not null"`
	MessageID int64     `json:"message_id" gorm:"column:message_id;type:bigint unsigned;not null;index"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;type:bigint unsigned;not null;index"`
	Answer    Answer    `json:"answer" gorm:"column:answer;type:json;not null"`
	Status    int8      `json:"status" gorm:"column:status;type:tinyint unsigned;not null;default:0"`
	Score     *float64  `json:"score" gorm:"column:score;type:decimal(5,2)"`
	Comment   string    `json:"comment" gorm:"column:comment;type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

type Answer struct {
	Choice    []int8   `json:"choice,omitempty"`
	TrueFalse bool     `json:"true_false,omitempty"`
	Text      string   `json:"text,omitempty"`
	Score     *float64 `json:"score,omitempty"`
}
