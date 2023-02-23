package entity

import "time"

type UserInfo struct {
	ID              int64  `json:"id"`               // 用户id
	UserName        string `json:"name"`             // 用户名称
	Avatar          string `json:"avatar"`           // 用户头像
	BackgroundImage string `json:"background_image"` // 用户个人页顶部大图
	Signature       string `json:"signature"`        // 个人简介
	//TotalFavorited  string `json:"total_favorited"`  // 获赞数量
	//WorkCount       int64  `json:"work_count"`       // 作品数
	//FavoriteCount   int64  `json:"favorite_count"`   // 喜欢数
	//FollowCount     int64  `json:"follow_count"`     // 关注总数
	//FollowerCount   int64  `json:"follower_count"`   // 粉丝总数
	CreatedAt time.Time `json:"created-at"` // 用户创建时间
	UpdatedAt time.Time `json:"update_at"`  // 用户上次更新用户数据的时间
}

func (UserInfo) TableName() string {
	return "user_info"
}
