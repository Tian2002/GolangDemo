package entity

import "time"

type Video struct {
	ID       int64  `json:"id"`        // 视频唯一标识
	UserId   int64  `json:"user_id"`   // 作者id
	Title    string `json:"title"`     // 视频标题
	PlayURL  string `json:"play_url"`  // 视频播放地址
	CoverURL string `json:"cover_url"` // 视频封面地址
	//FavoriteCount int64  `json:"favorite_count"` // 视频的点赞总数
	//CommentCount  int64  `json:"comment_count"`  // 视频的评论总数
	CreatedAt time.Time `json:"created-at"` // 视频发布时间
}

func (Video) TableName() string {
	return "videos"
}
