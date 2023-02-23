package vo

import "bytedance_douyin/biz/model/entity"

//UserInfo 视频作者信息
type UserInfo struct {
	entity.UserInfo
	IsFollow       bool  `json:"is_follow"`
	TotalFavorited int64 `json:"total_favorited"` // 获赞数量
	WorkCount      int64 `json:"work_count"`      // 作品数
	FavoriteCount  int64 `json:"favorite_count"`  // 喜欢数
	FollowCount    int64 `json:"follow_count"`    // 关注总数
	FollowerCount  int64 `json:"follower_count"`  // 粉丝总数
}
