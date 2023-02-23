package handler

import (
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/server"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"strconv"
)

func CommentAction(ctx context.Context, c *app.RequestContext) {
	u, _ := c.Get("userId")
	userId, _ := strconv.ParseInt(fmt.Sprintf("%v", u), 10, 64)
	videoId, _ := strconv.ParseInt(c.Query("video_id"), 10, 64)
	action := c.Query("action_type")
	context := c.Query("comment_text")
	commentId, _ := strconv.ParseInt(c.Query("comment_id"), 10, 64)
	commentInfo, err := server.CommentAction(userId, videoId, commentId, action, context)
	if err != nil {
		c.JSON(consts.StatusOK, vo.Comment{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "操作失败",
			},
		})
		log.Println("CommentAction", err.Error())
		return
	}
	c.JSON(consts.StatusOK, vo.Comment{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "操作成功",
		},
		Comment: &commentInfo,
	})
}

func CommentList(ctx context.Context, c *app.RequestContext) {
	//u, _ := c.Get("userId")
	//userId, _ := strconv.ParseInt(fmt.Sprintf("%v", u), 10, 64)

	commentList, err := server.CommentList(c.Query("video_id"))
	if err != nil {
		commentList.StatusCode = -1
		commentList.StatusMsg = "返回评论列表失败" + err.Error()
		c.JSON(consts.StatusOK, commentList)
		log.Println("CommentList", err.Error())
		return
	}
	commentList.StatusCode = 0
	commentList.StatusMsg = "返回评论列表成功"
	c.JSON(consts.StatusOK, commentList)
}
