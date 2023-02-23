package dao

import (
	"bytedance_douyin/biz/model/entity"
	"errors"
)

func FavoriteAction(userId, videoId, videoUserId int64) error {
	if DB.Where("user_id=?", userId).Where("video_id=?", videoId).Find(&[]entity.Favorite{}).RowsAffected == 1 {
		return errors.New("请不要重复操作")
	}
	tx := DB.Create(&entity.Favorite{
		UserID:      userId,
		VideoID:     videoId,
		VideoUserID: videoUserId,
	})
	return tx.Error
}

func UnFavoriteAction(userId, videoId, videoUserId int64) error {
	tx := DB.Where("user_id=?", userId).Where("video_id=?", videoId).Delete(&entity.Favorite{})
	return tx.Error
}
