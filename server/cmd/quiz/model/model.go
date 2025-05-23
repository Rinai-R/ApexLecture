package model

type QuizJudge struct {
	Id     int64  `json:"id" gorm:"primary_key;type:bigint"`
	RoomId int64  `json:"room_id" gorm:"not null;type:bigint"`
	UserId int64  `json:"user_id" gorm:"not null;type:bigint"`
	Title  string `json:"title" gorm:"not null;type:text"`
	Answer bool   `json:"answer" gorm:"type:boolean;not null"`
}

type QuizChoice struct {
	Id     int64    `json:"id" gorm:"primary_key;type:bigint"`
	RoomId int64    `json:"quiz_id" gorm:"not null;type:bigint"`
	UserId int64    `json:"user_id" gorm:"not null;type:bigint"`
	Title  string   `json:"title" gorm:"not null;type:text"`
	Option []string `json:"option" gorm:"type:json;not null"`
	Answer []int8   `json:"answer" gorm:"type:json;not null"`
}

type QuizStatus struct {
	QuestionId  int64   `json:"question_id"`
	RoomId      int64   `json:"room_id"`
	RequiredNum int64   `json:"required_num"`
	CurrentNum  int64   `json:"current_num"`
	AcceptRate  float64 `json:"start_time"`
}
