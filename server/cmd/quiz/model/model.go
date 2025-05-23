package model

import (
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

type QuizJudge struct {
	Id     int64  `json:"id" gorm:"primary_key;type:bigint"`
	RoomId int64  `json:"room_id" gorm:"not null;type:bigint"`
	UserId int64  `json:"user_id" gorm:"not null;type:bigint"`
	Title  string `json:"title" gorm:"not null;type:text"`
	Answer bool   `json:"answer" gorm:"type:boolean;not null"`
}

type QuizChoice struct {
	Id     int64   `json:"id" gorm:"primary_key;type:bigint"`
	RoomId int64   `json:"quiz_id" gorm:"not null;type:bigint"`
	UserId int64   `json:"user_id" gorm:"not null;type:bigint"`
	Title  string  `json:"title" gorm:"not null;type:text"`
	Option Options `json:"option" gorm:"type:json;not null"`
	Answer Answers `json:"answer" gorm:"type:json;not null"`
}

type QuizStatus struct {
	QuestionId  int64   `json:"question_id"`
	RoomId      int64   `json:"room_id"`
	RequiredNum int64   `json:"required_num"`
	CurrentNum  int64   `json:"current_num"`
	AcceptRate  float64 `json:"start_time"`
}

type Options []string

type Answers []int8

func (o Options) Value() (driver.Value, error) {
	return sonic.Marshal(o)
}

func (a Answers) Value() (driver.Value, error) {
	return sonic.Marshal(a)
}
func (o *Options) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Options: expected []byte, got %T", src)
	}
	return sonic.Unmarshal(b, o)
}

func (a *Answers) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Answers: expected []byte, got %T", src)
	}
	return sonic.Unmarshal(b, a)
}
