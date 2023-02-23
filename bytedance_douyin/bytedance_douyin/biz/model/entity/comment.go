package entity

import "time"

type Comment struct {
	ID        int64
	VideoId   int64
	UserId    int64
	Context   string
	CreatedAt time.Time
}

func (Comment) TableName() string {
	return "comment"
}
