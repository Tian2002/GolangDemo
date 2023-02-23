package handler

import (
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/server"
	"bytedance_douyin/biz/utils"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Feed(ctx context.Context, c *app.RequestContext) {

	token := c.Query("token")
	userId, _ := utils.ParseToken(token)
	//if err1 != nil {
	//	c.JSON(consts.StatusOK, vo.VideoArray{
	//		Status: vo.Status{
	//			StatusCode: -1,
	//			StatusMsg:  "用户未登录",
	//		},
	//	})
	//	return
	//}

	videoResponseList, err := server.Feed(c.Query("latest_time"), userId)
	if err != nil {
		c.JSON(consts.StatusOK, vo.VideoArray{
			Status:    vo.Status{},
			NextTime:  time.Now().Unix(),
			VideoList: nil,
		})
		log.Println("Feed", err.Error())
		return
	}
	c.JSON(consts.StatusOK, vo.VideoArray{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "视频列表返回成功",
		},
		NextTime:  time.Now().Unix(),
		VideoList: videoResponseList,
	})
}

func PublishVideo(ctx context.Context, c *app.RequestContext) {
	value, _ := c.Get("userId")
	userId, _ := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
	videoTitle := c.PostForm("title")
	videoData, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusOK, vo.Status{
			StatusCode: -1,
			StatusMsg:  "视频传输失败",
		})
		log.Println("PublishVideo", err.Error())
		return
	}
	err = server.PublishVideo(userId, videoTitle, videoData, c)
	if err != nil {
		c.JSON(http.StatusOK, vo.Status{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
		log.Println("PublishVideo", err.Error())
	} else {
		c.JSON(http.StatusOK, vo.Status{
			StatusCode: 0,
			StatusMsg:  "视频上传成功",
		})

	}
}

func PublishList(ctx context.Context, c *app.RequestContext) {
	userId := c.Query("user_id")
	value, _ := c.Get("userId")
	myId := fmt.Sprintf("%v", value)
	list, err := server.PublishList(userId, myId)
	if err != nil {
		c.JSON(consts.StatusOK, vo.VideoArray{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "视频列表获取失败",
			},
		})
		log.Println("PublishList", err.Error())
		return
	}
	c.JSON(consts.StatusOK, vo.VideoArray{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "返回列表成功",
		},
		VideoList: list,
	})
}
