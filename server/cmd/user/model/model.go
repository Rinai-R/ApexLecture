package model

type User struct {
	ID       int64  `json:"id" gorm:"type:bigint;column:id;primary_key;unique"`
	Username string `json:"username" gorm:"type:varchar(30);column:username;unique;not null;index:idx_username_password"`
	Password string `json:"password" gorm:"type:varchar(30);column:password;not null;index:idx_username_password"`
}
