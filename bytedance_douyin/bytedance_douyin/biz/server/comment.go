package server

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/utils"
	"strconv"
)

func CommentAction(userId, videoId, commentId int64, action, context string) (vo.CommentInfo, error) {
	var commentInfo vo.CommentInfo
	actionType, err := utils.ActionValid(action)
	if err != nil {
		return commentInfo, err
	}
	comment := &entity.Comment{
		VideoId: videoId,
		UserId:  userId,
		Context: context,
	}
	if actionType { //添加
		err = dao.CreateComment(comment)
		if err != nil {
			return commentInfo, err
		}
		commentInfo.ID = comment.ID
		userInfo := &vo.UserInfo{}
		err = dao.GetUserInfoById(userId, userInfo)
		if err != nil {
			return commentInfo, err
		}
		commentInfo.User = *userInfo
		commentInfo.Content = context
		commentInfo.CreateDate = comment.CreatedAt.Format("01-02")
	} else { //删除的情况下直接返回commentInfo为null
		err = dao.DeleteComment(commentId)
		if err != nil {
			return commentInfo, err
		}
	}

	return commentInfo, err
}

func CommentList(v string) (vo.CommentList, error) {
	videoId, _ := strconv.ParseInt(v, 10, 64)
	return dao.GetCommentListByVideoId(videoId)

}
