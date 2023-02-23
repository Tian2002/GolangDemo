package entity

import "time"

type User struct {
	ID        int64     `json:"id"`         // 用户id
	UserName  string    `json:"user_name"`  // 用户名
	Password  string    `json:"password"`   // 用户密码
	CreatedAt time.Time `json:"created-at"` // 用户创建时间
	UpdatedAt time.Time `json:"update_at"`  // 用户上次更新用户数据的时间
}

func (User) TableName() string {
	return "users"
}
