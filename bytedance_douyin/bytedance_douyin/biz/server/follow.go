package server

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/utils"
	"errors"
)

func FollowAction(userId, toUserId int64, actionType string) error {
	action, err := utils.ActionValid(actionType)
	if err != nil {
		return err
	}
	if action { //关注
		TorF, err2 := IsFollow(userId, toUserId)
		if err2 != nil {
			return err2
		}
		if TorF == true {
			return errors.New("已关注")
		}
		return dao.Follow(userId, toUserId)
	} else { //取消关注
		return dao.UnFollow(userId, toUserId)
	}

}
