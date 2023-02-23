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

var count = 0

func FavoriteAction(ctx context.Context, c *app.RequestContext) {
	count++
	if err := server.FavoriteAction(c.Query("token"), c.Query("action_type"), c.Query("video_id"), count); err != nil {
		c.JSON(consts.StatusOK, vo.Status{
			StatusCode: -1,
			StatusMsg:  "操作失败",
		})
		log.Println("FavoriteAction", err.Error())
		return
	}
	c.JSON(consts.StatusOK, vo.Status{
		StatusCode: 0,
		StatusMsg:  "操作成功",
	})
}

func FavoriteList(ctx context.Context, c *app.RequestContext) {
	u, _ := c.Get("userId")
	userId, _ := strconv.ParseInt(fmt.Sprintf("%v", u), 10, 64)
	//println(userId)
	list, err := server.FavoriteList(userId)
	if err != nil {
		c.JSON(consts.StatusOK, vo.VideoArray{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "获取视频列表失败" + err.Error(),
			},
		})
		log.Println("FavoriteList", err.Error())
		return
	}
	c.JSON(consts.StatusOK, vo.VideoArray{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "获取视频列表成功",
		},
		VideoList: list,
	})
}
