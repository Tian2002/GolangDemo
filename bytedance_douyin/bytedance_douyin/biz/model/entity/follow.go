package entity

import "time"

type Follow struct {
	ID        int64
	UserId    int64 //关注者
	FollowId  int64 //被关注者
	CreatedAt time.Time
}

func (Follow) TableName() string {
	return "follow"
}
