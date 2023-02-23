package dao

import (
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"errors"
	"gorm.io/gorm"
)

var DB *gorm.DB = new(gorm.DB)

func QueryUser(user *entity.User) error {
	if DB.Where("user_name=?", user.UserName).Find(&[]entity.User{}).RowsAffected == 1 { //优化
		return errors.New("用户名重复")
	}
	return nil
}

func AddInUser(user *entity.User, db *gorm.DB) error {

	//存信息到User表
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	//存储信息到关注表，自己对自己是关注的
	//var follow entity.Follow
	//follow.UserId = user.ID
	//follow.FollowId = user.ID
	//if err := db.Create(&follow).Error; err != nil {
	//	return err
	//}

	return nil
}
func AddInUserInfo(userInfo *entity.UserInfo, db *gorm.DB) error {
	//存信息到UserInfo表
	if err := db.Create(&userInfo).Error; err != nil {
		return err
	}
	return nil
}

func Login(user *entity.User) error {
	users := &[]entity.User{}
	DB.Where("user_name=? and password=?", user.UserName, user.Password).Find(users)
	if len(*users) != 1 {
		return errors.New("用户名或密码错误")
	}
	*user = (*users)[0]
	return nil
}

func GetUserInfoById(userId int64, userInfo *vo.UserInfo) error {
	err := DB.Where("id=?", userId).Find(&userInfo).Error
	if err != nil {
		return err
	}
	return nil
}

func GetTotalFavorited(userId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Favorite{}).Where("video_user_id=?", userId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil

}

func GetWorkCount(userId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Video{}).Where("user_id=?", userId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
func GetFavoriteCount(userId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Favorite{}).Where("user_id=?", userId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetFollowCount(userId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Follow{}).Where("user_id=?", userId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
func GetFollowerCount(userId int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.Follow{}).Where("follow_id=?", userId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func IsFavorite(userId int64, videoId int64) (bool, error) {
	if userId == videoId {
		return false, nil
	}
	var isFavorite entity.Favorite
	err := DB.Model(&entity.Favorite{}).Where("user_id=? and video_id=?", userId, videoId).Find(&isFavorite).Error
	if err != nil {
		return false, err
	}
	if isFavorite.ID != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func IsFollow(userId int64, toUserId int64) (bool, error) {
	var isFollow entity.Follow
	err := DB.Model(&entity.Follow{}).Where("user_id=? and follow_id=?", userId, toUserId).Find(&isFollow).Error
	if err != nil {
		return false, err
	}
	if isFollow.ID != 0 {
		return true, nil
	} else {
		return false, nil
	}
}
