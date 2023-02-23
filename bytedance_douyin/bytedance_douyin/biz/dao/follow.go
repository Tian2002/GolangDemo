package dao

import "bytedance_douyin/biz/model/entity"

func Follow(userId, toUserId int64) error {
	return DB.Create(&entity.Follow{
		UserId:   userId,
		FollowId: toUserId,
	}).Error
}

func UnFollow(userId, toUserId int64) error {
	return DB.Where("user_id=?", userId).Where("follow_id=?", toUserId).Delete(entity.Follow{}).Error
}
