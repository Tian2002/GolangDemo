package vo

type FavoriteList struct {
	Status
	VideoList []VideoInfo `json:"video_list"` // 用户点赞视频列表
}
