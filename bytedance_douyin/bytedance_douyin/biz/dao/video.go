package dao

import (
	"bytedance_douyin/biz/model/entity"
)

const MaxVideos = 30

func GetVideosLimitMaxVideos(videoList *[]entity.Video) error {
	err := DB.Model(&entity.Video{}).Order("created_at desc").Limit(MaxVideos).Find(&videoList).Error
	if err != nil {
		return err
	}
	return nil
}

func GetVideosLimitMaxVideosByLatestTime(videoList *[]entity.Video, latestTime int64) error { //需要改gorm
	err := DB.Model(&entity.Video{}).Order("created_at desc").Limit(MaxVideos).Find(&videoList)
	if err != nil {
		return err.Error
	}
	return nil
}

func CreateVideoByEntity(video *entity.Video) error {
	return DB.Create(video).Error
}
func GetVideoFavoriteCount(videoId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Favorite{}).Where("video_id=?", videoId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
func GetVideoCommentCount(videoId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Comment{}).Where("video_id=?", videoId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil

}
func GetVideoUserIdByVideoId(videoId int64) (int64, error) {
	var videoStruct entity.Video
	err := DB.Model(&entity.Video{}).Where("id=?", videoId).Find(&videoStruct).Error
	if err != nil {
		return 0, err
	}
	return videoStruct.UserId, nil
}

func GetVideoListByUserId(userId int64, videoList *[]entity.Video) error {
	return DB.Where("user_id=?", userId).Find(videoList).Error
}
