package handler

import (
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/server"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"log"
	"net/http"
	"strconv"
)

func FollowAction(ctx context.Context, c *app.RequestContext) {
	u, _ := c.Get("userId")
	userId, _ := strconv.ParseInt(fmt.Sprintf("%v", u), 10, 64)
	toUserId, _ := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	err := server.FollowAction(userId, toUserId, c.Query("action_type"))
	//if err.Error() == "已关注" {
	//	c.JSON(http.StatusOK, vo.Status{
	//		StatusCode: -1,
	//		StatusMsg:  err.Error(),
	//	})
	//	return
	//}
	if err != nil {
		c.JSON(http.StatusOK, vo.Status{
			StatusCode: -1,
			StatusMsg:  "操作失败",
		})
		log.Println("FollowAction", err.Error())
		return
	}
	c.JSON(http.StatusOK, vo.Status{
		StatusCode: 0,
	})

}
