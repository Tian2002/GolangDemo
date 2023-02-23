package dao

import (
	"bytedance_douyin/biz/model/entity"
	"bytedance_douyin/biz/model/vo"
	"fmt"
)

func CreateComment(comment *entity.Comment) error {
	return DB.Create(comment).Error
}

func GetCommentListByVideoId(videoId int64) (vo.CommentList, error) {
	commentList := vo.CommentList{}
	comments := make([]entity.Comment, 0)
	err := GetCommentsByVideoId(videoId, &comments)

	if err != nil {
		return commentList, err
	}
	for _, v := range comments {
		userInfo := &vo.UserInfo{}
		err = GetUserInfoById(v.UserId, userInfo)
		if err != nil {
			return commentList, err
		}
		commentInfo := vo.CommentInfo{
			Content:    v.Context,
			CreateDate: v.CreatedAt.Format("01-02"),
			ID:         v.ID,
			User:       *userInfo,
		}
		commentList.CommentList = append(commentList.CommentList, commentInfo)
	}

	return commentList, err
}

func GetCommentsByVideoId(videoId int64, comments *[]entity.Comment) error {
	r := DB.Where("video_id=?", videoId).Find(comments)

	fmt.Println(comments)

	return r.Error
}

func DeleteComment(commentId int64) error {
	return DB.Where("id=?", commentId).Delete(&entity.Comment{}).Error
}
