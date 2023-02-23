package handler

import (
	"bytedance_douyin/biz/model/vo"
	"bytedance_douyin/biz/server"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"net/http"
	"strconv"
)

type UserBaseInfo struct {
	vo.Status
	UserId int64  `json:"user_id"`
	Token  string `json:"token"`
}

func Register(ctx context.Context, c *app.RequestContext) {
	j := UserBaseInfo{
		Status: vo.Status{},
		UserId: 0,
		Token:  "",
	}
	password, exists := c.Get("password")
	if !exists {
		j.StatusCode = -1
		j.StatusMsg = "注册失败"
		c.JSON(consts.StatusOK, j)
		log.Println("Register", "取出密码失败")
		return
	}
	userId, token, err := server.Register(c.Query("username"), fmt.Sprintf("%v", password))
	if err != nil {
		j.StatusCode = -1
		j.StatusMsg = "注册失败"
		c.JSON(consts.StatusOK, j)
		log.Println("Register", err.Error())
		return
	}
	j.StatusCode = 0
	j.StatusMsg = "注册成功"
	j.UserId = userId
	j.Token = token
	c.JSON(consts.StatusOK, j)
}

func Login(ctx context.Context, c *app.RequestContext) {
	username := c.Query("username")
	pswd, _ := c.Get("password")
	password := fmt.Sprintf("%v", pswd)

	//转到server层处理
	userId, token, err := server.Login(username, password)
	if err != nil {
		c.JSON(http.StatusOK, UserBaseInfo{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "登陆失败",
			},
		})
		log.Println("Login", err.Error())
		return
	}
	c.JSON(http.StatusOK, UserBaseInfo{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "登陆成功",
		},
		UserId: userId,
		Token:  token,
	})

}

type UserInfoResp struct {
	vo.Status
	User vo.UserInfo `json:"user"`
}

func UserInfo(ctx context.Context, c *app.RequestContext) {
	t, _ := c.Get("token")
	token := fmt.Sprintf("%v", t)

	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, UserInfoResp{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "用户名格式错误",
			},
		})
		log.Println("UserInfo", err.Error())
	}

	userInfo, err1 := server.GetUserInfo(userId, token)
	if err1 != nil {
		c.JSON(http.StatusOK, UserInfoResp{
			Status: vo.Status{
				StatusCode: -1,
				StatusMsg:  "查询用户信息失败",
			},
		})
		log.Println("UserInfo", err.Error())
		return
	}
	c.JSON(consts.StatusOK, UserInfoResp{
		Status: vo.Status{
			StatusCode: 0,
			StatusMsg:  "查询用户信息成功",
		},
		User: userInfo,
	})

}
